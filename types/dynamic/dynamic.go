package dynamic

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// AttrGetter is required to be implemented by types that need to work with
// GetField and Marshal.
type AttrGetter interface {
	GetExtendedAttributes() []byte
}

// AttrSetter is required to be implemented by types that need to work with
// Unmarshal.
type AttrSetter interface {
	SetExtendedAttributes([]byte)
}

// Attributes is a combination of AttrGetter and AttrSetter.
type Attributes interface {
	AttrSetter
	AttrGetter
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
func SetField(v Attributes, path string, value interface{}) error {
	strukt := reflect.Indirect(reflect.ValueOf(v))
	if !strukt.IsValid() {
		return errors.New("SetField on nil Attributes")
	}
	if kind := strukt.Kind(); kind != reflect.Struct {
		return fmt.Errorf("invalid type (want struct): %v", kind)
	}
	addressOfAttrs := addressOfExtendedAttributes(v)
	fieldsp := structFieldPool.Get().(*[]structField)
	defer func() {
		*fieldsp = (*fieldsp)[:0]
		structFieldPool.Put(fieldsp)
	}()
	getJSONFields(strukt, addressOfAttrs, true, fieldsp)
	fields := *fieldsp
	f, ok := lookupField(fields, path)
	if !ok {
		return setExtendedAttribute(v, path, value)
	}
	field := f.Value
	if !field.IsValid() {
		return setExtendedAttribute(v, path, value)
	}
	if !field.CanSet() {
		return fmt.Errorf("dynamic: can't set struct field %q", path)
	}
	field.Set(reflect.ValueOf(value))
	return nil
}

// setExtendedAttributes inserts a value into v according to path. path is a
// dot-separated path like "foo.bar.baz".
//
// If setExtendedAttributes finds a path that does not currently exist, it will
// call makeEnvelope in order to create the necessary objects are value to
// satisfy the path.
//
// The mechanism of how this works is basically:
// 1) Lazy-unmarshal extended attributes (keys that reference []byte)
// 2) Write the extended attributes to a stream
//    a) If the key we are writing matches the first component of the path,
//       then deserialize the value into map[string]interface{} and insert
//       the value. Marshal the result and write it to the stream.
//    b) Otherwise, write the key-value pair as-is to the stream.
// 3) Set the extended attributes from the stream's buffer.
func setExtendedAttribute(v Attributes, path string, value interface{}) error {
	parts := strings.Split(strings.TrimSpace(path), ".")
	attrs := v.GetExtendedAttributes()
	any := jsoniter.Get(attrs)
	stream := jsoniter.NewStream(jsoniter.ConfigCompatibleWithStandardLibrary, nil, 1024)
	stream.WriteObjectStart()
	i := 0
	keys := any.Keys()
	sort.Strings(keys)
	for _, key := range keys {
		if key == parts[0] {
			continue
		}
		if i > 0 {
			stream.WriteMore()
		}
		stream.WriteObjectField(key)
		any := any.Get(key)
		any.WriteTo(stream)
		i++
	}
	if i > 0 {
		stream.WriteMore()
	}
	stream.WriteObjectField(parts[0])
	if len(parts) == 1 {
		stream.WriteVal(value)
	} else {
		envelope := makeEnvelope(any.Get(parts[0]), parts[1:], value)
		stream.WriteVal(envelope)
	}

	stream.WriteObjectEnd()
	if err := stream.Error; err != nil {
		return err
	}
	v.SetExtendedAttributes(stream.Buffer())
	return nil
}

// GetField gets a field from v according to its name.
// If GetField doesn't find a struct field with the corresponding name, then
// it will try to dynamically find the corresponding item in the 'Extended'
// field. GetField is case-sensitive, but extended attribute names will be
// converted to CamelCaps.
func GetField(v AttrGetter, name string) (interface{}, error) {
	if len(name) == 0 {
		return nil, errors.New("dynamic: empty path specified")
	}
	if v == nil {
		return nil, errors.New("dynamic: GetField with nil AttrGetter")
	}

	extendedAttributesAddress := addressOfExtendedAttributes(v)

	if s := string([]rune(name)[0]); strings.Title(s) == s {
		// Exported fields are always upper-cased for the first rune
		strukt := reflect.Indirect(reflect.ValueOf(v))
		if kind := strukt.Kind(); kind != reflect.Struct {
			return nil, fmt.Errorf("invalid type (want struct): %v", kind)
		}
		field := strukt.FieldByName(name)
		if field.IsValid() {
			rval := reflect.Indirect(field).Interface()
			if b, ok := rval.([]byte); ok && len(b) > 0 {
				// Make sure this field isn't the extended attributes
				if extendedAttributesAddress == &b[0] {
					goto EXTENDED
				}
			}
			return rval, nil
		}
	}
EXTENDED:
	// If we get here, we are dealing with extended attributes.
	any := AnyParameters{any: jsoniter.Get(v.GetExtendedAttributes())}
	return any.Get(name)
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

// Synthesize constructs a map[string]interface{} using the provided v in order
// to provide all the extended attributes as well as any fields in the concrete
// type of v
func Synthesize(v AttrGetter) (map[string]interface{}, error) {
	out := map[string]interface{}{}

	value := reflect.Indirect(reflect.ValueOf(v))
	t := value.Type()

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, received %s", t.Kind().String())
	}

	extendedAttributesAddress := addressOfExtendedAttributes(v)

	for i := 0; i < value.NumField(); i++ {
		field := t.Field(i)
		s := structField{Field: field}
		_, omitEmpty := s.jsonFieldName()

		// Don't add empty/nil fields to the map if omitempty is specified
		empty := isEmpty(value.Field(i))
		if empty && omitEmpty {
			continue
		}

		// Determine if we are handling custom attributes
		if !empty && isExtendedAttributes(extendedAttributesAddress, value.Field(i)) {
			var attrs interface{}
			if err := json.Unmarshal(value.Field(i).Bytes(), &attrs); err != nil {
				return nil, err
			}

			extendedAttributes := mapOfExtendedAttributes(attrs)
			for k, v := range extendedAttributes {
				out[k] = v
			}

			continue
		}

		out[field.Name] = value.Field(i).Interface()
	}

	return out, nil
}
