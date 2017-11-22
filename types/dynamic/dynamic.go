package dynamic

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// CustomAttributer is for use with GetField. It allows GetField to access
// serialized custom attributes.
type CustomAttributer interface {
	// CustomAttributes returns json-serialized custom attributes.
	CustomAttributes() []byte
}

// GetField gets a field from v according to its name.
// If GetField doesn't find a struct field with the corresponding name, then
// it will try to dynamically find the corresponding item in the 'Custom'
// field. GetField is case-sensitive, but custom attribute names will be
// converted to CamelCaps.
func GetField(v CustomAttributer, name string) (interface{}, error) {
	strukt := reflect.Indirect(reflect.ValueOf(v))
	if kind := strukt.Kind(); kind != reflect.Struct {
		return nil, fmt.Errorf("invalid type (want struct): %v", kind)
	}
	fields := getFields(strukt)
	field, ok := fields[name]
	if ok {
		return field.Value.Interface(), nil
	}
	// If we get here, we are dealing with custom attributes.
	return getCustomAttribute(v.CustomAttributes(), name)
}

// getCustomAttribute dynamically builds a concrete type. If the concrete
// type is a composite type, then it will either be a struct or a slice.
func getCustomAttribute(msg []byte, name string) (interface{}, error) {
	any := jsoniter.Get(msg, name)
	if err := any.LastError(); err != nil {
		lowerName := fmt.Sprintf("%s%s", strings.ToLower(string(name[0])), name[1:])
		if name != lowerName {
			// fall back to lower-case name
			return getCustomAttribute(msg, lowerName)
		}
		return nil, err
	}
	if any.GetInterface() == nil {
		// Fall back to lower-case name
		lowerName := fmt.Sprintf("%s%s", strings.ToLower(string(name[0])), name[1:])
		any = jsoniter.Get(msg, lowerName)
	}
	value, err := anyToValue(any)
	return value, err
}

func anyToValue(any jsoniter.Any) (interface{}, error) {
	switch any.ValueType() {
	case jsoniter.InvalidValue:
		return nil, fmt.Errorf("dynamic: %s", any.LastError())
	case jsoniter.StringValue:
		return any.ToString(), nil
	case jsoniter.NumberValue:
		return any.ToFloat64(), nil
	case jsoniter.NilValue:
		return nil, nil
	case jsoniter.BoolValue:
		return any.ToBool(), nil
	case jsoniter.ArrayValue:
		return buildSliceAny(any)
	case jsoniter.ObjectValue:
		return buildStructAny(any)
	default:
		return nil, fmt.Errorf("dynamic: unrecognized value type! %d", any.ValueType())
	}
}

// buildSliceAny dynamically builds a slice from a jsoniter.Any
func buildSliceAny(any jsoniter.Any) (interface{}, error) {
	n := any.Size()
	result := make([]interface{}, 0, n)
	for i := 0; i < n; i++ {
		value, err := anyToValue(any.Get(i))
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}

// buildStructAny dynamically builds a struct from a jsoniter.Any
func buildStructAny(any jsoniter.Any) (interface{}, error) {
	keys := any.Keys()
	fields := make([]reflect.StructField, 0, len(keys))
	values := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		value, err := anyToValue(any.Get(key))
		if err != nil {
			return nil, err
		}
		values = append(values, value)
		fields = append(fields, reflect.StructField{
			Name: strings.Title(key),
			Type: reflect.TypeOf(value),
		})
	}
	structType := reflect.StructOf(fields)
	structPtr := reflect.New(structType)
	structVal := reflect.Indirect(structPtr)
	for i, value := range values {
		field := structVal.Field(i)
		field.Set(reflect.ValueOf(value))
	}

	return structVal.Interface(), nil
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
	zeroValue := reflect.Zero(s.Field.Type).Interface()
	return reflect.DeepEqual(zeroValue, s.Value.Interface())
}

func (s structField) jsonFieldName() (string, bool) {
	fieldName := s.Field.Name
	tag, ok := s.Field.Tag.Lookup("json")
	omitEmpty := false
	if ok {
		parts := strings.Split(tag, ",")
		fieldName = parts[0]
		if len(parts) > 1 && parts[1] == "omitempty" {
			omitEmpty = true
		}
	}
	return fieldName, omitEmpty
}

func getJSONFields(v reflect.Value) map[string]structField {
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
		// sf is a valid JSON field.
		result[sf.JSONName] = sf
	}
	return result
}

// ExtractCustomAttributes selects only custom attributes from msg. It will
// ignore any fields in msg that correspond to fields in v. v must be of kind
// reflect.Struct.
func ExtractCustomAttributes(v interface{}, msg []byte) ([]byte, error) {
	strukt := reflect.Indirect(reflect.ValueOf(v))
	if kind := strukt.Kind(); kind != reflect.Struct {
		return nil, fmt.Errorf("invalid type (want struct): %v", kind)
	}
	fields := getJSONFields(strukt)
	stream := jsoniter.NewStream(jsoniter.ConfigDefault, nil, 4096)
	var anys map[string]jsoniter.Any
	if err := jsoniter.Unmarshal(msg, &anys); err != nil {
		return nil, err
	}
	objectStarted := false
	for _, any := range sortAnys(anys) {
		_, ok := fields[any.Name]
		if ok {
			// Not a custom attribute
			continue
		}
		if !objectStarted {
			objectStarted = true
			stream.WriteObjectStart()
		} else {
			stream.WriteMore()
		}
		stream.WriteObjectField(any.Name)
		any.WriteTo(stream)
	}
	if !objectStarted {
		stream.WriteObjectStart()
	}
	stream.WriteObjectEnd()
	return stream.Buffer(), nil
}

// Marshal encodes the struct fields in v that are valid to encode.
// It also encodes any custom attributes that are defined. Marshal
// respects the encoding/json rules regarding exported fields, and tag
// semantics. If v's kind is not reflect.Struct, an error will be returned.
func Marshal(v CustomAttributer) ([]byte, error) {
	s := jsoniter.NewStream(jsoniter.ConfigDefault, nil, 4096)
	s.WriteObjectStart()

	if err := encodeStructFields(v, s); err != nil {
		return nil, err
	}

	custom := v.CustomAttributes()
	if len(custom) > 0 {
		if err := encodeCustomFields(custom, s); err != nil {
			return nil, err
		}
	}

	s.WriteObjectEnd()

	return s.Buffer(), nil
}

func encodeStructFields(v interface{}, s *jsoniter.Stream) error {
	strukt := reflect.Indirect(reflect.ValueOf(v))
	if kind := strukt.Kind(); kind != reflect.Struct {
		return fmt.Errorf("invalid type (want struct): %v", kind)
	}
	m := getJSONFields(strukt)
	fields := make([]structField, 0, len(m))
	for _, s := range m {
		fields = append(fields, s)
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].JSONName < fields[j].JSONName
	})
	objectStarted := false
	for _, field := range fields {
		if !objectStarted {
			objectStarted = true
		} else {
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

func encodeCustomFields(custom []byte, s *jsoniter.Stream) error {
	var anys map[string]jsoniter.Any
	if err := jsoniter.Unmarshal(custom, &anys); err != nil {
		return err
	}
	for _, any := range sortAnys(anys) {
		s.WriteMore()
		s.WriteObjectField(any.Name)
		any.WriteTo(s)
	}
	return nil
}
