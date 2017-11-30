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
	if kind := strukt.Kind(); kind != reflect.Struct {
		return fmt.Errorf("invalid type (want struct): %v", kind)
	}
	attrs := v.GetExtendedAttributes()
	var addressOfAttrs *byte
	if len(attrs) > 0 {
		addressOfAttrs = &attrs[0]
	}
	fields := getJSONFields(strukt, addressOfAttrs)
	f, ok := fields[path]
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

// extractNonPathValues finds all the values in any that do not correspond to
// the path specified by parts.
func extractNonPathValues(any jsoniter.Any, parts []string) map[string]interface{} {
	keys := any.Keys()
	sort.Strings(keys)
	result := make(map[string]interface{}, len(keys))
	for _, key := range keys {
		result[key] = any.Get(key).GetInterface()
	}
	return result
}

// makeEnvelope makes an envelope of map[string]interface{} around any,
// according to parts. The nesting depth will be equal to the length of parts.
func makeEnvelope(any jsoniter.Any, parts []string, value interface{}) map[string]interface{} {
	remainingParts := parts
	result := extractNonPathValues(any, parts)
	envelope := result
	for len(remainingParts) > 1 {
		part := remainingParts[0]
		remainingParts = remainingParts[1:]
		var newEnv map[string]interface{}
		env, ok := envelope[part]
		if !ok {
			env = map[string]interface{}{}
		}
		if e, ok := env.(map[string]interface{}); !ok {
			newEnv = map[string]interface{}{}
		} else {
			newEnv = e
		}
		envelope[part] = newEnv
		envelope = newEnv
	}
	for i, part := range remainingParts {
		if i == len(remainingParts)-1 {
			envelope[part] = value
			break
		}
		m := make(map[string]interface{})
		envelope[part] = m
		envelope = m
	}
	return result
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
	extendedAttributes := v.GetExtendedAttributes()
	extAttrPtr := &extendedAttributes[0]
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
				if extAttrPtr == &b[0] {
					goto EXTENDED
				}
			}
			return rval, nil
		}
	}
EXTENDED:
	// If we get here, we are dealing with extended attributes.
	any := AnyParameters{any: jsoniter.Get(extendedAttributes)}
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

// structField is an internal convenience type
type structField struct {
	Field     reflect.StructField
	Value     reflect.Value
	JSONName  string
	OmitEmpty bool
}

func (s structField) IsEmpty() bool {
	zeroValue := reflect.Zero(reflect.Indirect(s.Value).Type()).Interface()
	return reflect.DeepEqual(zeroValue, s.Value.Interface())
}

func (s structField) jsonFieldName() (string, bool) {
	fieldName := s.Field.Name
	tag, ok := s.Field.Tag.Lookup("json")
	omitEmpty := false
	if ok {
		parts := strings.Split(tag, ",")
		if len(parts[0]) > 0 {
			fieldName = parts[0]
		}
		if len(parts) > 1 && parts[1] == "omitempty" {
			omitEmpty = true
		}
	}
	return fieldName, omitEmpty
}

func getJSONFields(v reflect.Value, addressOfAttrs *byte) map[string]structField {
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
		if sf.JSONName == "-" {
			continue
		}
		if sf.OmitEmpty && sf.IsEmpty() {
			continue
		}
		if b, ok := reflect.Indirect(value).Interface().([]byte); ok {
			if len(b) > 0 && &b[0] == addressOfAttrs {
				continue
			}
		}
		// sf is a valid JSON field.
		result[sf.JSONName] = sf
	}
	return result
}

// extractExtendedAttributes selects only extended attributes from msg. It will
// ignore any fields in msg that correspond to fields in v. v must be of kind
// reflect.Struct.
func extractExtendedAttributes(v interface{}, msg []byte) ([]byte, error) {
	strukt := reflect.Indirect(reflect.ValueOf(v))
	if kind := strukt.Kind(); kind != reflect.Struct {
		return nil, fmt.Errorf("invalid type (want struct): %v", kind)
	}
	fields := getJSONFields(strukt, nil)
	stream := jsoniter.NewStream(jsoniter.ConfigCompatibleWithStandardLibrary, nil, 4096)
	var anys map[string]jsoniter.Any
	if err := jsoniter.Unmarshal(msg, &anys); err != nil {
		return nil, err
	}
	j := 0
	for _, any := range sortAnys(anys) {
		_, ok := fields[any.Name]
		if ok {
			// Not a extended attribute
			continue
		}
		if j > 0 {
			stream.WriteMore()
		} else {
			stream.WriteObjectStart()
		}
		j++
		stream.WriteObjectField(any.Name)
		any.WriteTo(stream)
	}
	if j > 0 {
		stream.WriteObjectEnd()
	}
	buf := stream.Buffer()
	if len(buf) == 0 {
		buf = nil
	}
	return buf, nil
}

// Unmarshal decodes msg into v, storing what fields it can into the basic
// fields of the struct, and storing the rest into its extended attributes.
func Unmarshal(msg []byte, v AttrSetter) error {
	if _, ok := v.(json.Unmarshaler); ok {
		// Can't safely call UnmarshalJSON here without potentially causing an
		// infinite recursion. Copy the struct into a new type that doesn't
		// implement the method.
		oldVal := reflect.Indirect(reflect.ValueOf(v))
		typ := oldVal.Type()
		numField := typ.NumField()
		fields := make([]reflect.StructField, 0, numField)
		for i := 0; i < numField; i++ {
			field := typ.Field(i)
			if len(field.PkgPath) == 0 {
				fields = append(fields, field)
			}
		}
		newType := reflect.StructOf(fields)
		newPtr := reflect.New(newType)
		newVal := reflect.Indirect(newPtr)
		if err := json.Unmarshal(msg, newPtr.Interface()); err != nil {
			return err
		}
		for _, field := range fields {
			oldVal.FieldByName(field.Name).Set(newVal.FieldByName(field.Name))
		}
	} else {
		if err := json.Unmarshal(msg, v); err != nil {
			return err
		}
	}

	attrs, err := extractExtendedAttributes(v, msg)
	if err != nil {
		return err
	}
	if len(attrs) > 0 {
		v.SetExtendedAttributes(attrs)
	}
	return nil
}

// Marshal encodes the struct fields in v that are valid to encode.
// It also encodes any extended attributes that are defined. Marshal
// respects the encoding/json rules regarding exported fields, and tag
// semantics. If v's kind is not reflect.Struct, an error will be returned.
func Marshal(v AttrGetter) ([]byte, error) {
	s := jsoniter.NewStream(jsoniter.ConfigCompatibleWithStandardLibrary, nil, 4096)
	s.WriteObjectStart()

	if err := encodeStructFields(v, s); err != nil {
		return nil, err
	}

	extended := v.GetExtendedAttributes()
	if len(extended) > 0 {
		if err := encodeExtendedFields(extended, s); err != nil {
			return nil, err
		}
	}

	s.WriteObjectEnd()

	return s.Buffer(), nil
}

func encodeStructFields(v AttrGetter, s *jsoniter.Stream) error {
	strukt := reflect.Indirect(reflect.ValueOf(v))
	if kind := strukt.Kind(); kind != reflect.Struct {
		return fmt.Errorf("invalid type (want struct): %v", kind)
	}
	attrs := v.GetExtendedAttributes()
	var addressOfAttrs *byte
	if len(attrs) == 0 {
		addressOfAttrs = nil
	} else {
		addressOfAttrs = &attrs[0]
	}
	m := getJSONFields(strukt, addressOfAttrs)
	fields := make([]structField, 0, len(m))
	for _, s := range m {
		fields = append(fields, s)
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].JSONName < fields[j].JSONName
	})
	for i, field := range fields {
		if i > 0 {
			s.WriteMore()
		}
		s.WriteObjectField(field.JSONName)
		s.WriteVal(field.Value.Interface())
	}
	return nil
}

type anyT struct {
	Name string
	jsoniter.Any
}

func sortAnys(m map[string]jsoniter.Any) []anyT {
	anys := make([]anyT, 0, len(m))
	for key, any := range m {
		anys = append(anys, anyT{Name: key, Any: any})
	}
	sort.Slice(anys, func(i, j int) bool {
		return anys[i].Name < anys[j].Name
	})
	return anys
}

func encodeExtendedFields(extended []byte, s *jsoniter.Stream) error {
	var anys map[string]jsoniter.Any
	if err := jsoniter.Unmarshal(extended, &anys); err != nil {
		return err
	}
	for _, any := range sortAnys(anys) {
		s.WriteMore()
		s.WriteObjectField(any.Name)
		any.WriteTo(s)
	}
	return nil
}
