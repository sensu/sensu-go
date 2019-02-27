package graphql

import (
	"fmt"
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/stretchr/testify/assert"
)

func TestUintParseLiteral(t *testing.T) {
	testCases := []struct {
		in  ast.Value
		out interface{}
	}{
		{&ast.StringValue{Value: "123"}, 123},
		{&ast.IntValue{Value: "123"}, 123},
		{&ast.IntValue{Value: "12345678901234567890"}, nil},
		{&ast.FloatValue{Value: "12.43"}, nil},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("with %+v", tc.in), func(t *testing.T) {
			resolver := unsignedIntegerImpl{}
			res := resolver.ParseLiteral(tc.in)

			assert.EqualValues(t, tc.out, res)
		})
	}
}

func TestUintParseValue(t *testing.T) {
	testCases := []struct {
		in  interface{}
		out interface{}
	}{
		{uint8(123), 123},
		{uint16(123), 123},
		{uint32(123), 123},
		{uint64(123), 0},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("with %#v", tc.in), func(t *testing.T) {
			resolver := unsignedIntegerImpl{}
			res := resolver.ParseValue(tc.in)

			assert.EqualValues(t, tc.out, res)
		})
	}
}
