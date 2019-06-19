package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
)

// MutatorFilters returns collection of filters used for matching resources.
func MutatorFilters() map[string]filter.Filter {
	filters := map[string]filter.Filter{}

	// merge global filters
	for k, f := range GlobalFilters {
		filters[k] = f
	}

	return filters
}
