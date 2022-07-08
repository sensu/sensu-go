package postgres

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/selector"
	"github.com/sensu/sensu-go/backend/store"
)

var getEventTmpl = template.Must(template.New("GetEvents").Parse(`SELECT last_ok, occurrences, occurrences_wm, history_ts, history_status, history_index, serialized
FROM events
WHERE
{{ .NamespaceCond }}
AND {{ .EntityCond }}
AND {{ .CheckCond }}
AND {{ .SelectorCond }}
ORDER BY ({{ .Ordering }}) {{ .OrderingDirection }}
{{ .LimitClause }}
{{ .OffsetClause }}
;`))

type getEventTmplData struct {
	NamespaceCond     string
	EntityCond        string
	CheckCond         string
	SelectorCond      string
	Ordering          string
	OrderingDirection string
	LimitClause       string
	OffsetClause      string
}

var countEventsTmpl = template.Must(template.New("CountEvents").Parse(`SELECT count(*)
FROM events
WHERE
{{ .NamespaceCond }}
AND {{ .EntityCond }}
AND {{ .CheckCond }}
AND {{ .SelectorCond }}
;`))

type argCounter struct {
	value int
}

func (q *argCounter) Next() int {
	q.value++
	return q.value
}

// CreateGetEventsQuery creates a query to get events based on the input
// arguments. It supports querying by any combination of namespace, entity,
// check, label and field selectors, and limit and offset.
//
// The function returns a string which is the query, a slice of arguments to
// feed to the query, and an error if one occurred.
func CreateGetEventsQuery(namespace, entity, check string, s *selector.Selector, pred *store.SelectionPredicate) (string, []interface{}, error) {
	data, args, err := getTemplateData(namespace, entity, check, s, pred)
	if err != nil {
		return "", nil, err
	}
	var buf bytes.Buffer
	if err := getEventTmpl.Execute(&buf, data); err != nil {
		return "", nil, err
	}
	return buf.String(), args, nil
}

func CreateCountEventsQuery(namespace string, sel *selector.Selector, pred *store.SelectionPredicate) (string, []interface{}, error) {
	data, args, err := getTemplateData(namespace, "", "", sel, pred)
	if err != nil {
		return "", nil, err
	}
	var buf bytes.Buffer
	if err := countEventsTmpl.Execute(&buf, data); err != nil {
		return "", nil, err
	}
	return buf.String(), args, nil
}

func getTemplateData(namespace, entity, check string, s *selector.Selector, pred *store.SelectionPredicate) (getEventTmplData, []interface{}, error) {
	data := getEventTmplData{
		NamespaceCond: "true",
		EntityCond:    "true",
		CheckCond:     "true",
		SelectorCond:  "true",
		Ordering:      "sensu_namespace, sensu_entity, sensu_check",
	}
	var ctr argCounter
	args := make([]interface{}, 0, 4)
	if namespace != "" {
		data.NamespaceCond = fmt.Sprintf("sensu_namespace = $%d", ctr.Next())
		args = append(args, namespace)
	}
	if entity != "" {
		data.EntityCond = fmt.Sprintf("sensu_entity = $%d", ctr.Next())
		args = append(args, entity)
	}
	if check != "" {
		data.CheckCond = fmt.Sprintf("sensu_check = $%d", ctr.Next())
		args = append(args, check)
	}
	if s != nil && len(s.Operations) > 0 {
		builder := NewEventSelectorSQLBuilder(s)
		sc, sargs, err := builder.GetSelectorCond(&ctr)
		if err != nil {
			return data, nil, err
		}
		args = append(args, sargs...)
		data.SelectorCond = sc
	}

	if pred != nil {
		switch pred.Ordering {
		case "":
			break

		case corev2.EventSortEntity:
			data.Ordering = "sensu_entity"

		case corev2.EventSortLastOk:
			data.Ordering = "last_ok"

		case corev2.EventSortSeverity:
			data.Ordering = "CASE WHEN status = 0 THEN 3 WHEN status = 1 THEN 1 WHEN status = 2 THEN 0 ELSE 2 END"

		case corev2.EventSortTimestamp:
			data.Ordering = "history_ts[CASE WHEN history_index - 1 > 0 THEN history_index - 1 ELSE array_length(history_ts, 1)::INTEGER END]"

		default:
			return data, nil, errors.New("unknown ordering requested")
		}

		if pred.Descending {
			data.OrderingDirection = "DESC"
		}

		limit, offset, err := getLimitAndOffset(pred)
		if err != nil {
			return data, nil, err
		}
		if !limit.Valid {
			data.LimitClause = "LIMIT ALL"
		} else {
			data.LimitClause = fmt.Sprintf("LIMIT $%d", ctr.Next())
			args = append(args, limit.Int64)
		}
		data.OffsetClause = fmt.Sprintf("OFFSET $%d", ctr.Next())
		args = append(args, offset)
	}
	return data, args, nil
}
