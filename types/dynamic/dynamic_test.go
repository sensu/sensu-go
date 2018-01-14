package dynamic

import (
	"fmt"
	"reflect"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/sensu/govaluate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getFields gets a map of struct fields by name from a reflect.Value
func getFields(v reflect.Value) map[string]structField {
	typ := v.Type()
	numField := v.NumField()
	result := make(map[string]structField, numField)
	for i := 0; i < numField; i++ {
		field := typ.Field(i)
		if len(field.PkgPath) != 0 {
			// unexported
			continue
		}
		value := v.Field(i)
		sf := structField{Field: field, Value: value}
		sf.JSONName, sf.OmitEmpty = sf.jsonFieldName()
		result[field.Name] = sf
	}
	return result
}

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

	Attrs []byte `json:",omitempty"` // note that this will not be marshalled directly, despite missing the `json"-"`!
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

func TestGetFieldEmptyBytes(t *testing.T) {
	m := MyType{
		Foo:   "hello",
		Attrs: []byte(``),
	}

	testCases := []struct {
		field     string
		expected  interface{}
		expectErr bool
	}{
		{
			field:     "Foo",
			expected:  "hello",
			expectErr: false,
		},
		{
			field:     "bar",
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.field, func(t *testing.T) {
			field, err := GetField(&m, tc.field)
			assert.Equal(t, tc.expected, field)
			if tc.expectErr {
				assert.Error(t, err)
			}
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
		_, _ = expr.Eval(m)
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

func TestNoLookupAttrsDirectly(t *testing.T) {
	m := MyType{
		Attrs: []byte(`{}`),
	}
	_, err := m.Get("Attrs")
	require.NotNil(t, err)
	assert.Equal(t, err.Error(), "[Attrs] not found")
}

func TestSetFieldOnStructField(t *testing.T) {
	var m MyType
	err := SetField(&m, "foo", "hello, world!")
	require.NoError(t, err)
	require.Equal(t, "hello, world!", m.Foo)
	require.Equal(t, len(m.GetExtendedAttributes()), 0)
}

func TestSetFieldOnExtendedAttributes(t *testing.T) {
	tests := []struct {
		Attrs    []byte
		Expected []byte
		Path     string
		Value    interface{}
	}{
		{
			Attrs:    []byte(`{}`),
			Expected: []byte(`{"extendedAttr":{"bar":42}}`),
			Path:     "extendedAttr.bar",
			Value:    42,
		},
		{
			Attrs:    []byte(`{"extendedAttr":{"bar":5,"baz":{"a":[1,2,3,4]}}}`),
			Expected: []byte(`{"extendedAttr":{"a":"value","bar":5,"baz":{"a":[1,2,3,4]}}}`),
			Path:     "extendedAttr.a",
			Value:    "value",
		},
		{
			Attrs:    []byte(`{"extendedAttr":{"bar":5,"baz":{"a":[1,2,3,4],"b":"b"}}}`),
			Expected: []byte(`{"extendedAttr":{"bar":5,"baz":{"a":"replaced","b":"b"}}}`),
			Path:     "extendedAttr.baz.a",
			Value:    "replaced",
		},
		{
			Attrs:    []byte(`{"extendedAttr":{"bar":5,"baz":{"a":[1,2,3,4],"b":"b"}}}`),
			Expected: []byte(`{"extendedAttr":{"bar":5,"baz":{"a":{"b":"replaced"},"b":"b"}}}`),
			Path:     "extendedAttr.baz.a.b",
			Value:    "replaced",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var m MyType
			m.SetExtendedAttributes(test.Attrs)
			err := SetField(&m, test.Path, test.Value)
			require.NoError(t, err)
			require.Equal(t, string(test.Expected), string(m.GetExtendedAttributes()))
		})
	}
}

func TestSynthesize(t *testing.T) {
	testCases := []struct {
		name     string
		input    AttrGetter
		expected map[string]interface{}
	}{
		{
			name:     "empty input",
			input:    &MyType{},
			expected: map[string]interface{}{"Bar": []MyType(nil), "Foo": ""},
		},
		{
			name:  "standard fields",
			input: &MyType{Foo: "bar"},
			expected: map[string]interface{}{
				"Bar": []MyType(nil),
				"Foo": "bar",
			},
		},
		{
			name: "extended fields",
			input: &MyType{
				Foo:   "bar",
				Attrs: []byte(`{"baz": "qux"}`),
			},
			expected: map[string]interface{}{
				"Bar": []MyType(nil),
				"Baz": "qux",
				"Foo": "bar",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _ := Synthesize(tc.input)
			assert.Equal(t, tc.expected, reflect.ValueOf(result).Interface())
		})
	}
}
