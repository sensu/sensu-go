package dynamic

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"

	jsoniter "github.com/json-iterator/go"
)

// Marshal encodes the struct fields in v that are valid to encode.
// It also encodes any extended attributes that are defined. Marshal
// respects the encoding/json rules regarding exported fields, and tag
// semantics. If v's kind is not reflect.Struct, an error will be returned.
func Marshal(v AttrGetter) ([]byte, error) {
	if v == nil || reflect.ValueOf(v).IsNil() {
		return []byte("null"), nil
	}
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

// Unmarshal decodes msg into v, storing what fields it can into the basic
// fields of the struct, and storing the rest into its extended attributes.
func Unmarshal(msg []byte, v AttrSetter) error {
	if _, ok := v.(json.Unmarshaler); ok {
		// Can't safely call UnmarshalJSON here without potentially causing an
		// infinite recursion. Copy the struct into a new type that doesn't
		// implement the method.
		oldVal := reflect.Indirect(reflect.ValueOf(v))
		if !oldVal.IsValid() {
			return errors.New("Unmarshal called with nil AttrSetter")
		}
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

func encodeStructFields(v AttrGetter, s *jsoniter.Stream) error {
	strukt := reflect.Indirect(reflect.ValueOf(v))
	if !strukt.IsValid() {
		// Zero value of a nil pointer, nothing to do.
		return nil
	}
	if kind := strukt.Kind(); kind != reflect.Struct {
		return fmt.Errorf("invalid type (want struct): %v", kind)
	}

	extendedAttributesAddress := addressOfExtendedAttributes(v)

	m := getJSONFields(strukt, extendedAttributesAddress)
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
		elem := reflect.Indirect(value)
		sf := structField{Field: field, Value: value}
		sf.JSONName, sf.OmitEmpty = sf.jsonFieldName()
		if sf.JSONName == "-" {
			continue
		}
		if sf.OmitEmpty && isEmpty(sf.Value) {
			continue
		}
		if elem.IsValid() {
			if isExtendedAttributes(addressOfAttrs, value) {
				continue
			}
		}
		// if the field is embedded, flatten it out
		if sf.Field.Anonymous {
			var attrAddr *byte
			if x, ok := sf.Value.Interface().(AttrGetter); ok {
				attrs := x.GetExtendedAttributes()
				if len(attrs) > 0 {
					attrAddr = &attrs[0]
				}
			}
			fields := getJSONFields(sf.Value, attrAddr)
			for k, v := range fields {
				result[k] = v
			}
			continue
		}
		// sf is a valid JSON field.
		result[sf.JSONName] = sf
	}
	return result
}
