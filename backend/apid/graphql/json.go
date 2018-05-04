package graphql

import "github.com/sensu/sensu-go/graphql"
import definition "github.com/graphql-go/graphql"
import "github.com/graphql-go/graphql/language/ast"

var _ graphql.ScalarResolver = (*jsonImpl)(nil)

// Implement ScalarResolver for JSON type

type jsonImpl struct{}

func (jsonImpl) ParseLiteral(val ast.Value) interface{} {
	switch val := val.(type) {
	case *ast.IntValue:
		return definition.Int.ParseLiteral(val)
	case *ast.FloatValue:
		return definition.Float.ParseLiteral(val)
	case *ast.StringValue:
		return definition.String.ParseLiteral(val)
	case *ast.BooleanValue:
		return definition.Boolean.ParseLiteral(val)
	}
	// NOTE: Object and List values are not implemented.
	panic("Parsing of object and list values are not supported at this time.")
}

func (jsonImpl) ParseValue(val interface{}) interface{} {
	return val
}

func (jsonImpl) Serialize(val interface{}) interface{} {
	return val
}

type jsonWrapper struct {
	data []byte
}

func (w jsonWrapper) MarshalJSON() ([]byte, error) {
	return w.data, nil
}

func wrapExtendedAttributes(j []byte) jsonWrapper {
	if len(j) == 0 {
		j = []byte("{}")
	}
	return jsonWrapper{data: j}
}
