package dynamic

import (
	"reflect"
	"testing"

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

type MyType struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

type MyTypeEmbedded struct {
	Foo         string `json:"foo"`
	Bar         int    `json:"bar"`
	NotEmbedded Meta   `json:"not-embed"`
	Meta        `json:"meta"`
}

type Meta struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

func (m *MyType) Get(name string) (interface{}, error) {
	return GetField(m, name)
}

func TestGetField(t *testing.T) {
	m := &MyType{
		Foo: "hello",
		Bar: 42,
	}

	tests := []struct {
		Field string
		Exp   interface{}
	}{
		{
			Field: "Foo",
			Exp:   "hello",
		},
		{
			Field: "Bar",
			Exp:   int(42),
		},
	}

	for _, test := range tests {
		t.Run(test.Field, func(t *testing.T) {
			field, err := GetField(m, test.Field)
			require.NoError(t, err)
			assert.Equal(t, test.Exp, field)
		})
	}
}

func TestSetFieldOnStructField(t *testing.T) {
	var m MyType
	err := SetField(&m, "foo", "hello, world!")
	require.NoError(t, err)
	require.Equal(t, "hello, world!", m.Foo)
}

func TestSynthesize(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected map[string]interface{}
	}{
		{
			name:     "empty input",
			input:    &MyType{},
			expected: map[string]interface{}{"bar": 0, "foo": ""},
		},
		{
			name:  "standard fields",
			input: &MyType{Foo: "bar", Bar: 5},
			expected: map[string]interface{}{
				"bar": 5,
				"foo": "bar",
			},
		},
		{
			name: "embedded fields",
			input: &MyTypeEmbedded{
				Foo: "bar",
				Bar: 5,
				Meta: Meta{
					Name:        "baz",
					Namespace:   "default",
					Labels:      map[string]string{"Hi": "hello"},
					Annotations: map[string]string{"One": "1", "Two": "2"},
				},
				NotEmbedded: Meta{
					Name:        "not-baz",
					Namespace:   "not-default",
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
			},
			expected: map[string]interface{}{
				"bar":       5,
				"foo":       "bar",
				"name":      "baz",
				"namespace": "default",
				"labels": map[string]interface{}{
					"Hi": "hello",
				},
				"annotations": map[string]interface{}{
					"One": "1",
					"Two": "2",
				},
				"not-embed": map[string]interface{}{
					"name":        "not-baz",
					"namespace":   "not-default",
					"labels":      map[string]interface{}{},
					"annotations": map[string]interface{}{},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Synthesize(tc.input)
			assert.Equal(t, tc.expected, reflect.ValueOf(result).Interface())
		})
	}
}
