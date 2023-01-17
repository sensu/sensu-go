package postgres

import (
	"testing"

	"github.com/sensu/sensu-go/backend/selector"
)

func TestGetSelectorCond(t *testing.T) {
	builder := &SelectorSQLBuilder{
		selectorColumn:  "testSelectorCol",
		nestedSelectors: true,
		labelColumn:     "testLabelCol",
		labelPrefixes:   []string{""},
	}

	testCases := []struct {
		input         string
		expectedQuery string
	}{
		{
			input:         "foo == bar",
			expectedQuery: "testSelectorCol @> $1",
		},
		{
			input:         "foo == bar && zip != zap && bim == bap",
			expectedQuery: "testSelectorCol @> $1 AND NOT testSelectorCol @> $2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			selector, err := selector.ParseFieldSelector(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			builder.selector = selector
			actualQuery, _, err := builder.GetSelectorCond(&argCounter{0})
			if err != nil {
				t.Error(err)
				return
			}
			if actualQuery != tc.expectedQuery {
				t.Errorf("expected %s, got %s", tc.expectedQuery, actualQuery)
			}
		})
	}

}
