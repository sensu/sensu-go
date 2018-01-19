package graphql

import "github.com/sensu/sensu-go/graphql"

// To help implement types that will not used in an interface or union and as
// such will never need to be resolved.
type noLookup struct{}

func (*noLookup) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	return false
}
