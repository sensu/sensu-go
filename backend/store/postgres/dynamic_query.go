package postgres

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/lib/pq"
	corev2 "github.com/sensu/core/v2"
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

var eventFieldKeys = map[string]struct{}{}

func init() {
	fields := corev2.EventFields(corev2.FixtureEvent("entity", "check"))
	for k := range fields {
		eventFieldKeys[k] = struct{}{}
	}
}

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
		sc, sargs, err := getSelectorCond(&ctr, s)
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

func getSelectorCond(ctr *argCounter, s *selector.Selector) (string, []interface{}, error) {
	vars := make([]interface{}, 0, 4)
	conds := make([]string, 0, 4)
	inclusions := map[string]string{}
	exclusions := map[string]string{}
	for _, op := range s.Operations {
		switch op.Operator {
		case selector.DoubleEqualSignOperator, selector.NotEqualOperator, selector.MatchesOperator:
			if len(op.RValues) != 1 {
				return "", nil, fmt.Errorf("invalid operator: %v", op)
			}
		}
		switch op.Operator {
		case selector.InOperator:
			cond, vr := inOperatorMatch(ctr, op)
			conds = append(conds, cond)
			vars = append(vars, vr...)
		case selector.DoubleEqualSignOperator:
			if op.OperationType == selector.OperationTypeLabelSelector {
				cond, vr := inOperatorMatch(ctr, op)
				conds = append(conds, cond)
				vars = append(vars, vr...)
			} else {
				inclusions[op.LValue] = op.RValues[0]
			}
		case selector.NotInOperator:
			cond, vr := notInOperatorMatch(ctr, op)
			conds = append(conds, cond)
			vars = append(vars, vr...)
		case selector.NotEqualOperator:
			if op.OperationType == selector.OperationTypeLabelSelector {
				cond, vr := notInOperatorMatch(ctr, op)
				conds = append(conds, cond)
				vars = append(vars, vr...)
			} else {
				exclusions[op.LValue] = op.RValues[0]
			}
		case selector.MatchesOperator:
			cond, vr := matchOperator(ctr, op)
			conds = append(conds, cond)
			vars = append(vars, vr...)
		default:
			return "", nil, fmt.Errorf("unsupported operator: %s", op.Operator)
		}
	}

	cnds, vrs := formatSelectorConds(ctr, inclusions, exclusions)
	conds = append(conds, cnds...)
	vars = append(vars, vrs...)

	return strings.Join(conds, " AND "), vars, nil
}

func setRegexp(rvalues []string) string {
	var builder strings.Builder
	_ = builder.WriteByte('(')

	for i, value := range rvalues {
		if i != 0 {
			_ = builder.WriteByte('|')
		}
		_ = builder.WriteByte('(')
		_, _ = builder.Write([]byte("(^|,)"))
		_, _ = builder.WriteString(value)
		_, _ = builder.Write([]byte("($|,)"))
		_ = builder.WriteByte(')')
	}

	_ = builder.WriteByte(')')

	return builder.String()
}

// inOperatorMatch tries to match the behaviour of the event filter selector API.
// it is therefore quite complex and has four distinct cases that need to be
// accounted for. refer to ./backend/apid/filters/selector for more information.
func inOperatorMatch(ctr *argCounter, op selector.Operation) (string, []interface{}) {
	if len(op.RValues) == 0 {
		return "false", nil
	}
	if _, ok := eventFieldKeys[op.LValue]; ok {
		// event.entity.name in [foo, bar]
		selectorArg := ctr.Next()
		matchArg := ctr.Next()
		query := fmt.Sprintf("selectors ? $%d AND selectors->>$%d ~ $%d", selectorArg, selectorArg, matchArg)
		return query, []interface{}{op.LValue, setRegexp(op.RValues)}
	} else if strings.HasPrefix(op.LValue, "labels.") || op.OperationType == selector.OperationTypeLabelSelector && !strings.HasPrefix(op.RValues[0], "labels.") {
		// labels.foo in [foo, bar]
		if !strings.HasPrefix(op.LValue, "labels.") {
			op.LValue = fmt.Sprintf("labels.%s", op.LValue)
		}
		keyArg := ctr.Next()
		matchArg := ctr.Next()
		queryFragment := "(selectors ? (%s || $%d) AND selectors->>(%s || $%d) ~ $%d)"
		fragments := make([]string, 0, 3)
		for _, prefix := range []string{"event.", "event.entity.", "event.check."} {
			pfx := pq.QuoteLiteral(prefix)
			fragments = append(fragments, fmt.Sprintf(queryFragment, pfx, keyArg, pfx, keyArg, matchArg))
		}
		query := strings.Join(fragments, " OR ")
		return query, []interface{}{op.LValue, setRegexp(op.RValues)}
	} else if strings.HasPrefix(op.RValues[0], "labels.") {
		// foo in labels.foobars
		keyArg := ctr.Next()
		matchArg := ctr.Next()
		queryFragment := "(selectors ? (%s || $%d) AND selectors->>(%s || $%d) ~ $%d)"
		fragments := make([]string, 0, 3)
		for _, prefix := range []string{"event.", "event.entity.", "event.check."} {
			pfx := pq.QuoteLiteral(prefix)
			fragments = append(fragments, fmt.Sprintf(queryFragment, pfx, keyArg, pfx, keyArg, matchArg))
		}
		query := strings.Join(fragments, " OR ")
		return query, []interface{}{op.RValues[0], setRegexp([]string{op.LValue})}
	} else {
		// foo in event.check.subscriptions
		selectorArg := ctr.Next()
		matchArg := ctr.Next()
		query := fmt.Sprintf("selectors ? $%d AND selectors->>$%d ~ $%d", selectorArg, selectorArg, matchArg)
		return query, []interface{}{op.RValues[0], setRegexp([]string{op.LValue})}
	}
}

func notInOperatorMatch(ctr *argCounter, op selector.Operation) (string, []interface{}) {
	query, values := inOperatorMatch(ctr, op)
	query = fmt.Sprintf("NOT (%s)", query)
	return query, values
}

func matchOperator(ctr *argCounter, op selector.Operation) (string, []interface{}) {
	if len(op.RValues) != 1 {
		panic(fmt.Sprintf("rvalue len = %d != 1", len(op.RValues)))
	}
	if op.OperationType != selector.OperationTypeLabelSelector {
		query := fmt.Sprintf("selectors->>$%d ~ $%d", ctr.Next(), ctr.Next())
		return query, []interface{}{op.LValue, op.RValues[0]}
	}
	keyArg := ctr.Next()
	matchArg := ctr.Next()
	queryFragment := "selectors ? (%s || $%d) AND selectors->>(%s || $%d) ~ $%d"
	fragments := make([]string, 0, 3)
	for _, prefix := range []string{"event.", "event.entity.", "event.check."} {
		pfx := pq.QuoteLiteral(prefix)
		fragments = append(fragments, fmt.Sprintf(queryFragment, pfx, keyArg, pfx, keyArg, matchArg))
	}
	query := strings.Join(fragments, " OR ")
	return query, []interface{}{op.LValue, op.RValues}
}

func formatSelectorConds(ctr *argCounter, inclusions, exclusions map[string]string) ([]string, []interface{}) {
	conds := make([]string, 0, 2)
	vars := make([]interface{}, 0, 2)
	if len(inclusions) > 0 {
		conds = append(conds, fmt.Sprintf("selectors @> $%d", ctr.Next()))
		b, _ := json.Marshal(inclusions)
		vars = append(vars, b)
	}
	if len(exclusions) > 0 {
		conds = append(conds, fmt.Sprintf("NOT selectors @> $%d", ctr.Next()))
		b, _ := json.Marshal(exclusions)
		vars = append(vars, b)
	}
	return conds, vars
}
