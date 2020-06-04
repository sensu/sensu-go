package dynamic

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// SynthesizeExtras is a type that wants to pass extra values to the Synthesize
// function. The key-value pairs will be included as-is without inspection by
// the Synthesize function. This is useful for populated synthesized values with
// functions or computed values.
type SynthesizeExtras interface {
	// SynthesizeExtras returns a map of extra values to include when passing
	// to Synthesize().
	SynthesizeExtras() map[string]interface{}
}

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
func reflectMapToMapParameters(v reflect.Value) interface{} {
	result := make(map[string]interface{})
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

// Synthesize recursively turns structs into map[string]interface{}
// values. It works on most datatypes. Synthesize panics if it is
// called on channels.
//
// Synthesize will use the json tag from struct fields to name map
// keys, if the json tag is present.
func Synthesize(v interface{}) interface{} {
	value := reflect.Indirect(reflect.ValueOf(v))
	switch value.Kind() {
	case reflect.Struct:
		result := synthesizeStruct(value)
		if m, ok := v.(SynthesizeExtras); ok {
			for k, v := range m.SynthesizeExtras() {
				result[k] = v
			}
		}
		return result
	case reflect.Slice, reflect.Array:
		return synthesizeSlice(value)
	case reflect.Map:
		return synthesizeMap(value)
	case reflect.Chan:
		panic("can't synthesize channel")
	case reflect.Invalid:
		// We got passed a nil
		return nil
	default:
		if value.CanInterface() {
			return value.Interface()
		}
		return nil
	}
}

func synthesizeSlice(value reflect.Value) interface{} {
	length := value.Len()
	slice := make([]interface{}, length)
	for i := 0; i < length; i++ {
		val := value.Index(i)
		var elt interface{} = nil
		if val.CanInterface() {
			elt = val.Interface()
		}
		slice[i] = Synthesize(elt)
	}
	return slice
}

func synthesizeMap(value reflect.Value) interface{} {
	typ := value.Type()
	keyT := typ.Key()
	if keyT.Kind() != reflect.String {
		// Maps without string keys are not supported
		return map[string]interface{}{}
	}
	length := value.Len()
	out := make(map[string]interface{}, length)
	for _, key := range value.MapKeys() {
		val := value.MapIndex(key)
		var elt interface{}
		if val.CanInterface() {
			elt = val.Interface()
		}
		out[key.Interface().(string)] = Synthesize(elt)
	}
	return out
}

func synthesizeStruct(value reflect.Value) map[string]interface{} {
	numField := value.NumField()
	out := make(map[string]interface{}, numField)
	t := value.Type()
	for i := 0; i < numField; i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			// unexported fields are not included in the synthesis
			continue
		}
		s := structField{Field: field}
		fieldName, omitEmpty := s.jsonFieldName()
		fieldValue := value.Field(i)

		// Don't add empty/nil fields to the map if omitempty is specified
		empty := isEmpty(fieldValue)
		if empty && omitEmpty {
			continue
		}

		switch fieldValue.Kind() {
		case reflect.Struct:
			// Recursively convert all fields to synthesized values
			fields := synthesizeStruct(fieldValue)

			// flatten embedded fields to the top level
			if t.Field(i).Anonymous {
				for k, v := range fields {
					out[k] = v
				}
			} else {
				out[fieldName] = fields
			}
		case reflect.Slice, reflect.Array:
			out[fieldName] = synthesizeSlice(fieldValue)
		case reflect.Map:
			out[fieldName] = synthesizeMap(fieldValue)
		default:
			if fieldValue.CanInterface() {
				out[fieldName] = Synthesize(fieldValue.Interface())
			} else {
				out[fieldName] = nil
			}
		}
	}

	return out
}
