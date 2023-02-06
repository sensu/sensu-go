package postgres

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/selector"
	"github.com/sensu/sensu-go/backend/store"
)

var getEventTmpl = template.Must(template.New("GetEvents").Parse(`
WITH ns AS (
	SELECT id FROM namespaces
	WHERE namespaces.name = $1
)
SELECT serialized
FROM events
FULL OUTER JOIN ns
ON events.namespace = ns.id
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

var countEventsTmpl = template.Must(template.New("CountEvents").Parse(`
WITH ns AS (
	SELECT id FROM namespaces
	WHERE name = $1
)
SELECT count(*)
FROM events
FULL OUTER JOIN ns
ON events.namespace = ns.id
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
		Ordering:      "namespace, entity_name, check_name",
	}
	var ctr argCounter
	args := make([]interface{}, 0, 4)
	data.NamespaceCond = fmt.Sprintf("(namespace = ns.id OR $%d = '')", ctr.Next())
	args = append(args, namespace)
	if entity != "" {
		data.EntityCond = fmt.Sprintf("entity_name = $%d", ctr.Next())
		args = append(args, entity)
	}
	if check != "" {
		data.CheckCond = fmt.Sprintf("check_name = $%d", ctr.Next())
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
			data.Ordering = "entity_name"

		case corev2.EventSortLastOk:
			data.Ordering = "(selectors ->> 'event.check.last_ok')::INTEGER"

		case corev2.EventSortSeverity:
			data.Ordering = "CASE WHEN selectors ->> 'event.check.status' = '0' THEN 3 WHEN selectors ->> 'event.check.status' = '1' THEN 1 WHEN selectors ->> 'event.check.status' = '2' THEN 0 ELSE 2 END"

		case corev2.EventSortTimestamp:
			data.Ordering = "(selectors ->> 'event.timestamp')::INTEGER"

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
