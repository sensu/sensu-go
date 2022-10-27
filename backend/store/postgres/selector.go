package postgres

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/selector"
)

var eventFieldKeys = map[string]struct{}{}

func init() {
	fields := corev2.EventFields(corev2.FixtureEvent("entity", "check"))
	for k := range fields {
		eventFieldKeys[k] = struct{}{}
	}
}

type SelectorSQLBuilder struct {
	selectorColumn      string
	nestedSelectors     bool
	labelColumn         string
	labelPrefixes       []string
	includeLabelCaption bool
	validFieldKeys      map[string]struct{}
	selector            *selector.Selector
}

func NewEventSelectorSQLBuilder(selector *selector.Selector) *SelectorSQLBuilder {
	return &SelectorSQLBuilder{
		selectorColumn:      "selectors",
		nestedSelectors:     false,
		labelColumn:         "selectors",
		labelPrefixes:       []string{"event.", "event.entity.", "event.check."},
		includeLabelCaption: true,
		validFieldKeys:      eventFieldKeys,
		selector:            selector,
	}
}

func NewConfigSelectorSQLBuilder(selector *selector.Selector) *SelectorSQLBuilder {
	return &SelectorSQLBuilder{
		selectorColumn:      "resource",
		nestedSelectors:     true,
		labelColumn:         "labels",
		labelPrefixes:       []string{""},
		includeLabelCaption: false,
		validFieldKeys:      nil,
		selector:            selector,
	}
}

func (s *SelectorSQLBuilder) GetSelectorCond(ctr *argCounter) (string, []interface{}, error) {
	vars := make([]interface{}, 0, 4)
	conds := make([]string, 0, 4)
	inclusions := map[string]string{}
	exclusions := map[string]string{}
	for _, op := range s.selector.Operations {
		switch op.Operator {
		case selector.DoubleEqualSignOperator, selector.NotEqualOperator, selector.MatchesOperator:
			if len(op.RValues) != 1 {
				return "", nil, fmt.Errorf("invalid operator: %v", op)
			}
		}
		switch op.Operator {
		case selector.InOperator:
			cond, vr := s.inOperatorMatch(ctr, op)
			conds = append(conds, cond)
			vars = append(vars, vr...)
		case selector.DoubleEqualSignOperator:
			if op.OperationType == selector.OperationTypeLabelSelector {
				cond, vr := s.inOperatorMatch(ctr, op)
				conds = append(conds, cond)
				vars = append(vars, vr...)
			} else {
				inclusions[op.LValue] = op.RValues[0]
			}
		case selector.NotInOperator:
			cond, vr := s.notInOperatorMatch(ctr, op)
			conds = append(conds, cond)
			vars = append(vars, vr...)
		case selector.NotEqualOperator:
			if op.OperationType == selector.OperationTypeLabelSelector {
				cond, vr := s.notInOperatorMatch(ctr, op)
				conds = append(conds, cond)
				vars = append(vars, vr...)
			} else {
				exclusions[op.LValue] = op.RValues[0]
			}
		case selector.MatchesOperator:
			cond, vr := s.matchOperator(ctr, op)
			conds = append(conds, cond)
			vars = append(vars, vr...)
		default:
			return "", nil, fmt.Errorf("unsupported operator: %s", op.Operator)
		}
	}

	cnds, vrs := s.formatSelectorConds(ctr, inclusions, exclusions)
	conds = append(conds, cnds...)
	vars = append(vars, vrs...)

	return strings.Join(conds, " AND "), vars, nil
}

// inOperatorMatch tries to match the behaviour of the event filter selector API.
// it is therefore quite complex and has four distinct cases that need to be
// accounted for. refer to ./backend/apid/filters/selector for more information.
func (s *SelectorSQLBuilder) inOperatorMatch(ctr *argCounter, op selector.Operation) (string, []interface{}) {
	if len(op.RValues) == 0 {
		return "false", nil
	}
	//if _, ok := eventFieldKeys[op.LValue]; ok {
	if s.validFieldKey(&op) {
		// event.entity.name in [foo, bar]
		selectorArg := ctr.Next()
		matchArg := ctr.Next()
		//query := fmt.Sprintf("%s ? $%d AND %s->>$%d ~ $%d", s.selectorColumn, selectorArg, s.selectorColumn, selectorArg, matchArg)
		query := fmt.Sprintf("%s#>>$%d ~ $%d", s.selectorColumn, selectorArg, matchArg)
		return query, []interface{}{s.matchLValue(op.LValue), s.setRegexp(op.RValues)}
	} else if strings.HasPrefix(op.LValue, "labels.") || op.OperationType == selector.OperationTypeLabelSelector && !strings.HasPrefix(op.RValues[0], "labels.") {
		// labels.foo in [foo, bar]
		if !strings.HasPrefix(op.LValue, "labels.") && s.includeLabelCaption {
			op.LValue = fmt.Sprintf("labels.%s", op.LValue)
		}
		keyArg := ctr.Next()
		matchArg := ctr.Next()
		queryFragment := "(%s ? (%s || $%d) AND %s->>(%s || $%d) ~ $%d)"
		fragments := make([]string, 0, len(s.labelPrefixes))
		for _, prefix := range s.labelPrefixes {
			pfx := pq.QuoteLiteral(prefix)
			fragments = append(fragments, fmt.Sprintf(queryFragment, s.labelColumn, pfx, keyArg, s.labelColumn, pfx, keyArg, matchArg))
		}
		query := strings.Join(fragments, " OR ")
		return query, []interface{}{op.LValue, s.setRegexp(op.RValues)}
	} else if strings.HasPrefix(op.RValues[0], "labels.") {
		// foo in labels.foobars
		keyArg := ctr.Next()
		matchArg := ctr.Next()
		queryFragment := "(%s ? (%s || $%d) AND %s->>(%s || $%d) ~ $%d)"
		fragments := make([]string, 0, len(s.labelPrefixes))
		for _, prefix := range s.labelPrefixes {
			pfx := pq.QuoteLiteral(prefix)
			fragments = append(fragments, fmt.Sprintf(queryFragment, s.labelColumn, pfx, keyArg, s.labelPrefixes, pfx, keyArg, matchArg))
		}
		query := strings.Join(fragments, " OR ")
		return query, []interface{}{op.RValues[0], s.setRegexp([]string{op.LValue})}
	} else {
		// foo in event.check.subscriptions
		selectorArg := ctr.Next()
		matchArg := ctr.Next()
		query := fmt.Sprintf("%s ? $%d AND %s->>$%d ~ $%d", s.selectorColumn, selectorArg, s.selectorColumn, selectorArg, matchArg)
		return query, []interface{}{op.RValues[0], s.setRegexp([]string{op.LValue})}
	}
}

func (s *SelectorSQLBuilder) notInOperatorMatch(ctr *argCounter, op selector.Operation) (string, []interface{}) {
	query, values := s.inOperatorMatch(ctr, op)
	query = fmt.Sprintf("NOT (%s)", query)
	return query, values
}

func (s *SelectorSQLBuilder) setRegexp(rvalues []string) string {
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

func (s *SelectorSQLBuilder) matchOperator(ctr *argCounter, op selector.Operation) (string, []interface{}) {
	if len(op.RValues) != 1 {
		panic(fmt.Sprintf("rvalue len = %d != 1", len(op.RValues)))
	}
	if op.OperationType != selector.OperationTypeLabelSelector {
		query := fmt.Sprintf("%s#>>$%d ~ $%d", s.selectorColumn, ctr.Next(), ctr.Next())
		return query, []interface{}{s.matchLValue(op.LValue), op.RValues[0]}
	}
	keyArg := ctr.Next()
	matchArg := ctr.Next()
	queryFragment := "%s ? (%s || $%d) AND %s->>(%s || $%d) ~ $%d"
	fragments := make([]string, 0, 3)
	for _, prefix := range s.labelPrefixes {
		pfx := pq.QuoteLiteral(prefix)
		fragments = append(fragments, fmt.Sprintf(queryFragment, s.labelColumn, pfx, keyArg, s.labelColumn, pfx, keyArg, matchArg))
	}
	query := strings.Join(fragments, " OR ")
	return query, []interface{}{op.LValue, op.RValues[0]}
}

func (s *SelectorSQLBuilder) formatSelectorConds(ctr *argCounter, inclusions, exclusions map[string]string) ([]string, []interface{}) {
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

func (s *SelectorSQLBuilder) validFieldKey(op *selector.Operation) bool {
	if s.validFieldKeys == nil {
		return op.OperationType == selector.OperationTypeFieldSelector
	}
	_, ok := s.validFieldKeys[op.LValue]
	return ok
}

func (s *SelectorSQLBuilder) matchLValue(lValue string) string {
	if !s.nestedSelectors {
		return lValue
	}
	return fmt.Sprintf("{%s}", strings.ReplaceAll(lValue, ".", ","))
}
