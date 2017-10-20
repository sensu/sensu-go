package graphqlschema

import (
	"strconv"
	"time"

	"github.com/graphql-go/graphql"
	graphqlast "github.com/graphql-go/graphql/language/ast"
)

var timeScalar *graphql.Scalar

func init() {
	timeScalar = graphql.NewScalar(graphql.ScalarConfig{
		Name:        "Time",
		Description: "The `Time` scalar type represents an instant in time",
		Serialize:   coerceTime,
		ParseValue:  coerceTime,
		ParseLiteral: func(valueAST graphqlast.Value) interface{} {
			switch valueAST := valueAST.(type) {
			case *graphqlast.IntValue:
				if intValue, err := strconv.Atoi(valueAST.Value); err == nil {
					return time.Unix(int64(intValue), 0)
				}
			case *graphqlast.StringValue:
				// TODO: Would be nice to cover
			}
			return nil
		},
	})
}

func coerceTime(value interface{}) interface{} {
	switch value := value.(type) {
	case time.Time:
		return value.Format(time.RFC1123Z)
	case int64: // TODO: Too naive
		return coerceTime(time.Unix(value, 0))
	}

	return nil
}
