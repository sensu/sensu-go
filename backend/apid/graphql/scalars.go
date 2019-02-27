package graphql

import (
	"strconv"

	definition "github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/sensu/sensu-go/graphql"
)

var _ graphql.ScalarResolver = (*jsonImpl)(nil)
var _ graphql.ScalarResolver = (*unsignedIntegerImpl)(nil)

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

// Implement ScalarResolver for JSON type

type unsignedIntegerImpl struct{}

func (unsignedIntegerImpl) ParseLiteral(val ast.Value) interface{} {
	var inVal string
	switch val := val.(type) {
	case *ast.IntValue:
		inVal = val.Value
	case *ast.StringValue:
		inVal = val.Value
	}
	if len(inVal) > 0 {
		if v, err := parseUint(inVal); err == nil {
			return v
		}
	}
	return nil
}

func (unsignedIntegerImpl) ParseValue(val interface{}) interface{} {
	return coerceUint(val)
}

func (unsignedIntegerImpl) Serialize(val interface{}) interface{} {
	return coerceUint(val)
}

func parseUint(val string) (uint32, error) {
	uintVal, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(uintVal), nil
}

func coerceUint(value interface{}) interface{} {
	switch value := value.(type) {
	case uint8:
		return uint32(value)
	case *uint8:
		return coerceUint(value)
	case uint16:
		return uint32(value)
	case *uint16:
		return coerceUint(value)
	case uint32:
		return value
	case *uint32:
		return coerceUint(value)
	}
	return uint32(0)
}

// Implement ScalarResolver for BigNum type
