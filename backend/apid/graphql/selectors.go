package graphql

import (
	"strings"

	"github.com/sensu/sensu-go/backend/selector"
)

func parseEventFilters(statements []string) *selector.Selector {
	const statementSeparator = ":"
	var statusMap = map[string]string{
		"passing": "0",
		"warning": "1",
		"critial": "2",
	}
	selectors := make([]*selector.Selector, 0, len(statements))
	for _, s := range statements {
		ss := strings.SplitN(s, statementSeparator, 2)
		if len(ss) != 2 {
			logger.WithField("selector", ss).Warn("invalid selector")
			continue
		}
		switch ss[0] {
		case "fieldSelector":
			sel, err := selector.ParseFieldSelector(ss[1])
			if err != nil {
				logger.WithField("selector", ss).WithError(err).Warn("invalid selector")
				continue
			}
			selectors = append(selectors, sel)
		case "labelSelector":
			sel, err := selector.ParseLabelSelector(ss[1])
			if err != nil {
				logger.WithField("selector", ss).WithError(err).Warn("invalid selector")
				continue
			}
			selectors = append(selectors, sel)
		case "silenced":
			sel := &selector.Selector{
				Operations: []selector.Operation{
					{
						LValue:   "event.is_silenced",
						Operator: selector.DoubleEqualSignOperator,
						RValues:  []string{ss[1]},
					},
				},
			}
			selectors = append(selectors, sel)
		case "entity":
			sel := &selector.Selector{
				Operations: []selector.Operation{
					{
						LValue:   "event.entity.name",
						Operator: selector.DoubleEqualSignOperator,
						RValues:  []string{ss[1]},
					},
				},
			}
			selectors = append(selectors, sel)
		case "check":
			sel := &selector.Selector{
				Operations: []selector.Operation{
					{
						LValue:   "event.check.name",
						Operator: selector.DoubleEqualSignOperator,
						RValues:  []string{ss[1]},
					},
				},
			}
			selectors = append(selectors, sel)
		case "status":
			status, ok := statusMap[ss[1]]
			if !ok {
				if ss[1] == "incident" {
					sel := &selector.Selector{
						Operations: []selector.Operation{
							{
								LValue:   "event.check.status",
								Operator: selector.NotEqualOperator,
								RValues:  []string{"0"},
							},
						},
					}
					selectors = append(selectors, sel)
				}
				// we can't represent "incident" or "unknown" so avoid selecting
				// this altogether.
				continue
			}
			sel := &selector.Selector{
				Operations: []selector.Operation{
					{
						LValue:   "event.check.status",
						Operator: selector.DoubleEqualSignOperator,
						RValues:  []string{status},
					},
				},
			}
			selectors = append(selectors, sel)
		}
	}
	if len(selectors) == 0 {
		return nil
	}
	return selector.Merge(selectors...)
}
