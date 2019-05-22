package graphql

import (
	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
)

// HandlerFilters returns collection of filters used for matching resources.
func HandlerFilters() map[string]filter.Filter {
	filters := map[string]filter.Filter{
		// type:pipe | type:tcp | type:udp | type:set
		"type": filter.String(func(res v2.Resource, v string) bool {
			return res.(*v2.Handler).Type == v
		}),
	}

	// merge global filters
	for k, f := range GlobalFilters {
		filters[k] = f
	}

	return filters
}
