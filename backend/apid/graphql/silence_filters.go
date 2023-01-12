package graphql

import (
	v2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
)

// SilenceFilters returns collection of filters used for matching resources.
func SilenceFilters() map[string]filter.Filter {
	filters := map[string]filter.Filter{
		// check:disk-check
		"check": filter.String(func(res corev3.Resource, v string) bool {
			return res.(*v2.Silenced).Check == v
		}),

		// subscription:unix | subscription:db
		"subscription": filter.String(func(res corev3.Resource, v string) bool {
			return res.(*v2.Silenced).Subscription == v
		}),
	}

	// merge global filters
	for k, f := range GlobalFilters {
		filters[k] = f
	}

	return filters
}
