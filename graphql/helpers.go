package graphqlschema

import (
	"reflect"

	"github.com/graphql-go/graphql"
)

// AliasField TODO: ...
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
