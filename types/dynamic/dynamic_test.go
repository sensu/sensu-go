package dynamic

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Knetic/govaluate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStructField(t *testing.T) {
	assert := assert.New(t)

	test := struct {
		Valid             int `json:"valid"`
		invalidUnexported int
		ValidEmpty        int `json:"validEmpty"`
		InvalidEmpty      int `json:"invalidEmpty,omitempty"`
		Invalid           int `json:"-"`
	}{
		Valid:             5,
		invalidUnexported: 1,
		ValidEmpty:        0,
		InvalidEmpty:      0,
		Invalid:           10,
	}

	fields := getFields(reflect.ValueOf(test))
	require.Equal(t, len(fields), 4)

	field := fields["Valid"]
	assert.Equal(field.Value.Interface(), 5)
	assert.Equal("valid", field.JSONName)
	assert.Equal(false, field.OmitEmpty)

	field = fields["ValidEmpty"]
	assert.Equal(field.Value.Interface(), 0)
	assert.Equal("validEmpty", field.JSONName)
	assert.Equal(false, field.OmitEmpty)

	field = fields["InvalidEmpty"]
	assert.Equal(field.Value.Interface(), 0)

	field = fields["Invalid"]
	assert.Equal(field.Value.Interface(), 10)
}

func TestGetJSONStructField(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	test := struct {
		Valid             int `json:"valid"`
		invalidUnexported int
		ValidEmpty        int `json:"validEmpty"`
		InvalidEmpty      int `json:"invalidEmpty,omitempty"`
		Invalid           int `json:"-"`
	}{
		Valid:             5,
		invalidUnexported: 1,
		ValidEmpty:        0,
		InvalidEmpty:      0,
		Invalid:           10,
	}

	fields := getJSONFields(reflect.ValueOf(test))
	require.Equal(2, len(fields))

	field, ok := fields["valid"]
	require.Equal(true, ok)
	assert.Equal(field.Value.Interface(), 5)
	assert.Equal("valid", field.JSONName)
	assert.Equal(false, field.OmitEmpty)

	field = fields["validEmpty"]
	assert.Equal(field.Value.Interface(), 0)
	assert.Equal("validEmpty", field.JSONName)
	assert.Equal(false, field.OmitEmpty)
}

type MyType struct {
	Foo string   `json:"foo"`
	Bar []MyType `json:"bar"`

	custom []byte
}

func (m MyType) CustomAttributes() []byte {
	return m.custom
}

func (m MyType) Get(name string) (interface{}, error) {
	return GetField(m, name)
}

func (m MyType) MarshalJSON() ([]byte, error) {
	return Marshal(m)
}

func (m *MyType) UnmarshalJSON(p []byte) error {
	type __ MyType
	var x __
	if err := json.Unmarshal(p, &x); err != nil {
		return err
	}
	*m = MyType(x)
	custom, err := ExtractCustomAttributes(m, p)
	if err != nil {
		return err
	}
	m.custom = custom
	return nil
}

func TestExtractEmptyCustomAttributes(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	msg := []byte(`{"foo": "hello, world!","bar":[{"foo":"o hai"}]}`)
	var m MyType

	attrs, err := ExtractCustomAttributes(m, msg)
	require.Nil(err)
	assert.Equal([]byte("{}"), attrs)
}

func TestExtractCustomAttributes(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	msg := []byte(`{"foo": "hello, world!","bar":[{"foo":"o hai"}], "customattr": "such custom"}`)
	var m MyType

	attrs, err := ExtractCustomAttributes(m, msg)
	require.Nil(err)
	assert.Equal([]byte(`{"customattr":"such custom"}`), attrs)
}

func TestMarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	customBytes := []byte(`{"a":1,"b":2.0,"c":true,"d":"false","e":[1,2,3],"f":{"foo":"bar"}}`)
	expBytes := []byte(`{"bar":null,"foo":"hello world!","a":1,"b":2.0,"c":true,"d":"false","e":[1,2,3],"f":{"foo":"bar"}}`)

	m := MyType{
		Foo:    "hello world!",
		Bar:    nil,
		custom: customBytes,
	}

	b, err := Marshal(m)
	require.Nil(err)
	assert.Equal(expBytes, b)
}

func TestGetField(t *testing.T) {
	m := MyType{
		Foo:    "hello",
		Bar:    []MyType{{Foo: "there"}},
		custom: []byte(`{"a":"a","b":1,"c":2.0,"d":true,"e":null,"foo":{"hello":5},"bar":[true,10.5]}`),
	}

	tests := []struct {
		Field string
		Exp   interface{}
		Kind  reflect.Kind
	}{
		{
			Field: "a",
			Exp:   "a",
			Kind:  reflect.String,
		},
		{
			Field: "b",
			Exp:   1.0,
			Kind:  reflect.Float64,
		},
		{
			Field: "c",
			Exp:   2.0,
			Kind:  reflect.Float64,
		},
		{
			Field: "d",
			Exp:   true,
			Kind:  reflect.Bool,
		},
		{
			Field: "e",
			Exp:   nil,
			Kind:  reflect.Invalid,
		},
		{
			Field: "foo",
			Exp: struct {
				Hello float64
			}{Hello: 5.0},
			Kind: reflect.Struct,
		},
		{
			Field: "bar",
			Exp:   []interface{}{true, 10.5},
			Kind:  reflect.Slice,
		},
	}

	for _, test := range tests {
		t.Run(test.Field, func(t *testing.T) {
			field, err := GetField(m, test.Field)
			require.Nil(t, err)
			assert.Equal(t, test.Exp, field)
			assert.Equal(t, test.Kind, reflect.ValueOf(field).Kind())
		})
	}
}

func TestQueryGovaluateSimple(t *testing.T) {
	m := MyType{
		custom: []byte(`{"hello":5}`),
	}

	expr, err := govaluate.NewEvaluableExpression("hello == 5")
	require.Nil(t, err)
	require.NotNil(t, expr)

	result, err := expr.Eval(m)
	require.Nil(t, err)
	require.Equal(t, true, result)

	expr, err = govaluate.NewEvaluableExpression("Hello == 5")
	require.Nil(t, err)
	require.NotNil(t, expr)

	result, err = expr.Eval(m)
	require.Nil(t, err)
	require.Equal(t, true, result)
}

func TestQueryGovaluateComplex(t *testing.T) {
	m := MyType{
		custom: []byte(`{"hello":{"foo":5,"bar":6.0}}`),
	}

	expr, err := govaluate.NewEvaluableExpression("hello.Foo == 5")
	require.Nil(t, err)
	require.NotNil(t, expr)

	result, err := expr.Eval(m)
	require.Nil(t, err)
	require.Equal(t, true, result)

	expr, err = govaluate.NewEvaluableExpression("Hello.Foo == 5")
	require.Nil(t, err)
	require.NotNil(t, expr)

	result, err = expr.Eval(m)
	require.Nil(t, err)
	require.Equal(t, true, result)

	expr, err = govaluate.NewEvaluableExpression("Hello.Foo < Hello.Bar")
	require.Nil(t, err)
	require.NotNil(t, expr)

	result, err = expr.Eval(m)
	require.Nil(t, err)
	require.Equal(t, true, result)
}

func TestMarshalUnmarshal(t *testing.T) {
	data := []byte(`{"bar":null,"foo":"hello","a":10,"b":"c"}`)
	var m MyType
	err := json.Unmarshal(data, &m)
	require.Nil(t, err)
	assert.Equal(t, MyType{Foo: "hello", custom: []byte(`{"a":10,"b":"c"}`)}, m)
	b, err := json.Marshal(m)
	require.Nil(t, err)
	assert.Equal(t, data, b)
}
