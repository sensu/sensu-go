package graphqlschema

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

// AliasResolver makes it quick and easy to create a resolver that reflects a
// different field on a given resource.
//
// Usage:
//
// "legs": &graphql.Field{
//   Name:    "number of legs the owner's cat has",
//   Type:    graphql.Int,
//   Resolve: AliasResolver("myCat", "numberOfLegs"),
// },
//
func AliasResolver(fNames ...string) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		fVal := reflect.ValueOf(p.Source)
		for _, fName := range fNames {
			fVal = reflect.Indirect(fVal)
			fVal = fVal.FieldByName(fName)
		}
		return fVal.Interface(), nil
	}
}
