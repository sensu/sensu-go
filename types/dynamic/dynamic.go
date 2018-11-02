package dynamic

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// SetField inserts a value into v at path.
//
// For example, if the marshalled representation of v is
// {"foo": "bar", "baz": { "value": 5 }},
// Then SetField(v, "baz.value", 10) will result in
// {"foo": "bar", "baz": { "value": 10 }}.
//
// v's reflect.Kind must be reflect.Struct, or a non-nil error will
// be returned. If the path refers to a struct field, then v must
// be addressable, or an error will be returned.
func SetField(v interface{}, path string, value interface{}) error {
	strukt := reflect.Indirect(reflect.ValueOf(v))
	if !strukt.IsValid() {
		return errors.New("SetField on nil Attributes")
	}
	if kind := strukt.Kind(); kind != reflect.Struct {
		return fmt.Errorf("invalid type (want struct): %v", kind)
	}
	fieldsp := structFieldPool.Get().(*[]structField)
	defer func() {
		*fieldsp = (*fieldsp)[:0]
		structFieldPool.Put(fieldsp)
	}()
	getJSONFields(strukt, true, fieldsp)
	fields := *fieldsp
	f, ok := lookupField(fields, path)
	if !ok {
		return fmt.Errorf("dynamic: no such field: %q", path)
	}
	field := f.Value
	if !field.IsValid() {
		return fmt.Errorf("dynamic: field is invalid: %s", path)
	}
	if !field.CanSet() {
		return fmt.Errorf("dynamic: struct field not addressable: %q", path)
	}
	field.Set(reflect.ValueOf(value))
	return nil
}

// GetField gets a field from v according to its name.
// If GetField doesn't find a struct field with the corresponding name, then
// it will try to dynamically find the corresponding item in the 'Extended'
// field. GetField is case-sensitive, but extended attribute names will be
// converted to CamelCaps.
func GetField(v interface{}, name string) (interface{}, error) {
	if len(name) == 0 {
		return nil, errors.New("dynamic: empty path specified")
	}
	if v == nil {
		return nil, errors.New("dynamic: GetField with nil")
	}

	if s := string([]rune(name)[0]); strings.Title(s) != s {
		// Exported fields are always upper-cased for the first rune
		name = strings.Title(s) + string([]rune(name)[1:])
	}
	strukt := reflect.Indirect(reflect.ValueOf(v))
	if kind := strukt.Kind(); kind != reflect.Struct {
		return nil, fmt.Errorf("invalid type (want struct): %v", kind)
	}
	field := strukt.FieldByName(name)
	if field.IsValid() {
		field := reflect.Indirect(field)
		if field.Kind() == reflect.Map {
			return reflectMapToMapParameters(field), nil
		}
		return field.Interface(), nil
	}

	return nil, fmt.Errorf("missing field: %q", name)
}

// reflectMapToMapParameters turns a reflect.Map into a map[string]interface{}
// that implements govaluate.Parameters. A custom datatype is used here instead
// of govaluate.MapParameters, in order to avoid creating a package dependency
// here.
func reflectMapToMapParameters(v reflect.Value) interface{} {
	result := make(mapParameters)
	for _, key := range v.MapKeys() {
		if key.Kind() != reflect.String {
			// Fallback - if the map has a non-string key type, return the
			// variable as-is.
			return v.Interface()
		}
		result[key.Interface().(string)] = v.MapIndex(key).Interface()
	}
	return result
}

type mapParameters map[string]interface{}

func (m mapParameters) Get(name string) (interface{}, error) {
	v, ok := m[name]
	if !ok {
		return nil, fmt.Errorf("missing map key: %q", name)
	}
	return v, nil
}

// AnyParameters connects jsoniter.Any to govaluate.Parameters
type AnyParameters struct {
	any jsoniter.Any
}

// Get implements the govaluate.Parameters interface.
func (p AnyParameters) Get(name string) (interface{}, error) {
	any := p.any.Get(name)
	if err := any.LastError(); err != nil {
		return nil, err
	}
	switch any.ValueType() {
	case jsoniter.InvalidValue:
		return nil, fmt.Errorf("dynamic: %s", any.LastError())
	case jsoniter.ObjectValue:
		return AnyParameters{any: any}, any.LastError()
	default:
		return any.GetInterface(), any.LastError()
	}

}

// Synthesize constructs a map[string]interface{} from its input using reflection.
func Synthesize(v interface{}) (map[string]interface{}, error) {
	out := map[string]interface{}{}

	value := reflect.Indirect(reflect.ValueOf(v))
	t := value.Type()

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, received %s", t.Kind().String())
	}

	for i := 0; i < value.NumField(); i++ {
		field := t.Field(i)
		s := structField{Field: field}
		_, omitEmpty := s.jsonFieldName()

		// Don't add empty/nil fields to the map if omitempty is specified
		empty := isEmpty(value.Field(i))
		if empty && omitEmpty {
			continue
		}

		out[field.Name] = value.Field(i).Interface()
	}

	return out, nil
}
