package graphqlschema

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

// AliasField makes it quick and easy to create a field that reflects a
// different field on given resource.
//
// TODO: This should be replaced be a generic resolver and not a field; more flexibility.
func AliasField(T graphql.Output, fNames ...string) *graphql.Field {
	return &graphql.Field{
		Type: T,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			fVal := reflect.ValueOf(p.Source)
			for _, fName := range fNames {
				fVal = reflect.Indirect(fVal)
				fVal = fVal.FieldByName(fName)
			}
			return fVal.Interface(), nil
		},
	}
}
