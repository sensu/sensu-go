package graphql

import (
	v2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
)

// EventFilterFilters returns collection of filters used for matching Event Filter resources.
func EventFilterFilters() map[string]filter.Filter {
	filters := map[string]filter.Filter{
		// action:allow | action:deny
		"action": filter.String(func(res corev3.Resource, v string) bool {
			return res.(*v2.EventFilter).Action == v
		}),
	}

	// merge global filters
	for k, f := range GlobalFilters {
		filters[k] = f
	}

	return filters
}
