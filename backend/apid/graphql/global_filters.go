package graphql

import "github.com/sensu/sensu-go/backend/apid/graphql/filter"

// GlobalFilters are filters that are applied to all resolvers that accept
// filter statements.
var GlobalFilters = DefaultGlobalFilters()

// DefaultGlobalFilters returns the default set of global filters.
func DefaultGlobalFilters() map[string]filter.Filter {
	return map[string]filter.Filter{}
}
