package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.TimeWindowWhenFieldResolvers = (*timeWindowWhenImpl)(nil)
var _ schema.TimeWindowDaysFieldResolvers = (*timeWindowDaysImpl)(nil)
var _ schema.TimeWindowTimeRangeFieldResolvers = (*timeWindowTimeRangeImpl)(nil)

//
// Implement TimeWindowWhenFieldResolvers
//

type timeWindowWhenImpl struct {
	schema.TimeWindowWhenAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*timeWindowWhenImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(types.TimeWindowWhen)
	return ok
}

//
// Implement TimeWindowDaysFieldResolvers
//

type timeWindowDaysImpl struct {
	schema.TimeWindowDaysAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*timeWindowDaysImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(types.TimeWindowDays)
	return ok
}

//
// Implement TimeWindowTimeRangeFieldResolvers
//

type timeWindowTimeRangeImpl struct {
	schema.TimeWindowTimeRangeAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*timeWindowTimeRangeImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(types.TimeWindowTimeRange)
	return ok
}
