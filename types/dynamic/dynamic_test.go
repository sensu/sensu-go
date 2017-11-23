package dynamic

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Knetic/govaluate"
	jsoniter "github.com/json-iterator/go"
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
		Attributes        []byte
	}{
		Valid:             5,
		invalidUnexported: 1,
		ValidEmpty:        0,
		InvalidEmpty:      0,
		Invalid:           10,
		Attributes:        []byte(`"hello!"`),
	}

	fields := getJSONFields(reflect.ValueOf(test), &test.Attributes[0])
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

	Attrs []byte // note that this will not be marshalled directly, despite missing the `json"-"`!
}

func (m *MyType) GetExtendedAttributes() []byte {
	return m.Attrs
}

func (m *MyType) SetExtendedAttributes(a []byte) {
	m.Attrs = a
}

func (m *MyType) Get(name string) (interface{}, error) {
	return GetField(m, name)
}

func (m *MyType) MarshalJSON() ([]byte, error) {
	return Marshal(m)
}

func (m *MyType) UnmarshalJSON(p []byte) error {
	return Unmarshal(p, m)
}

func TestExtractEmptyExtendedAttributes(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	msg := []byte(`{"foo": "hello, world!","bar":[{"foo":"o hai"}]}`)
	var m MyType

	attrs, err := extractExtendedAttributes(m, msg)
	require.NoError(err)
	assert.Equal([]byte("{}"), attrs)
}

func TestExtractExtendedAttributes(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	msg := []byte(`{"foo": "hello, world!","bar":[{"foo":"o hai"}], "extendedattr": "such extended"}`)
	var m MyType

	attrs, err := extractExtendedAttributes(m, msg)
	require.NoError(err)
	assert.Equal([]byte(`{"extendedattr":"such extended"}`), attrs)
}

func TestMarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	extendedBytes := []byte(`{"a":1,"b":2.0,"c":true,"d":"false","e":[1,2,3],"f":{"foo":"bar"}}`)
	expBytes := []byte(`{"bar":null,"foo":"hello world!","a":1,"b":2.0,"c":true,"d":"false","e":[1,2,3],"f":{"foo":"bar"}}`)

	m := &MyType{
		Foo:   "hello world!",
		Bar:   nil,
		Attrs: extendedBytes,
	}

	b, err := Marshal(m)
	require.NoError(err)
	assert.Equal(expBytes, b)
}

func TestGetField(t *testing.T) {
	m := &MyType{
		Foo:   "hello",
		Bar:   []MyType{{Foo: "there"}},
		Attrs: []byte(`{"a":"a","b":1,"c":2.0,"d":true,"e":null,"foo":{"hello":5},"bar":[true,10.5]}`),
	}

	fooAny := jsoniter.Get([]byte(`{"hello":5}`))
	require.NoError(t, fooAny.LastError())

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
			Exp:   AnyParameters{any: fooAny},
			Kind:  reflect.Struct,
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
			require.NoError(t, err)
			assert.Equal(t, test.Exp, field)
			assert.Equal(t, test.Kind, reflect.ValueOf(field).Kind())
		})
	}
}

func TestQueryGovaluateSimple(t *testing.T) {
	m := &MyType{
		Attrs: []byte(`{"hello":5}`),
	}

	expr, err := govaluate.NewEvaluableExpression("hello == 5")
	require.NoError(t, err)
	require.NotNil(t, expr)

	result, err := expr.Eval(m)
	require.NoError(t, err)
	require.Equal(t, true, result)
}

func BenchmarkQueryGovaluateSimple(b *testing.B) {
	m := &MyType{
		Attrs: []byte(`{"hello":5}`),
	}

	expr, err := govaluate.NewEvaluableExpression("hello == 5")
	require.NoError(b, err)
	require.NotNil(b, expr)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		expr.Eval(m)
	}
}

func TestQueryGovaluateComplex(t *testing.T) {
	m := &MyType{
		Attrs: []byte(`{"hello":{"foo":5,"bar":6.0}}`),
	}

	expr, err := govaluate.NewEvaluableExpression("hello.foo == 5")
	require.NoError(t, err)
	require.NotNil(t, expr)

	result, err := expr.Eval(m)
	require.NoError(t, err)
	require.Equal(t, true, result)

	expr, err = govaluate.NewEvaluableExpression("hello.foo < hello.bar")
	require.NoError(t, err)
	require.NotNil(t, expr)

	result, err = expr.Eval(m)
	require.NoError(t, err)
	require.Equal(t, true, result)
}

func BenchmarkQueryGovaluateComplex(b *testing.B) {
	m := &MyType{
		Attrs: []byte(`{"hello":{"foo":5,"bar":6.0}}`),
	}

	expr, err := govaluate.NewEvaluableExpression("hello.foo < hello.bar")
	require.NoError(b, err)
	require.NotNil(b, expr)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = expr.Eval(m)
	}
}

func TestMarshalUnmarshal(t *testing.T) {
	data := []byte(`{"bar":null,"foo":"hello","a":10,"b":"c"}`)
	var m MyType
	err := json.Unmarshal(data, &m)
	require.NoError(t, err)
	assert.Equal(t, MyType{Foo: "hello", Attrs: []byte(`{"a":10,"b":"c"}`)}, m)
	b, err := json.Marshal(&m)
	require.NoError(t, err)
	assert.Equal(t, data, b)
}

func BenchmarkUnmarshal(b *testing.B) {
	data := []byte(`{"bar":null,"foo":"hello","a":10,"b":"c"}`)
	var m MyType
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(data, &m)
	}
}

func BenchmarkMarshal(b *testing.B) {
	data := []byte(`{"bar":null,"foo":"hello","a":10,"b":"c"}`)
	var m MyType
	json.Unmarshal(data, &m)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(&m)
	}
}

func TestNoLookupAttrsDirectly(t *testing.T) {
	m := MyType{
		Attrs: []byte(`{}`),
	}
	_, err := m.Get("Attrs")
	require.NotNil(t, err)
	assert.Equal(t, err.Error(), "[Attrs] not found")
}
