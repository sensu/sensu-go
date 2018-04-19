package dynamic

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

var (
	fastjson        = jsoniter.ConfigCompatibleWithStandardLibrary
	structFieldPool sync.Pool
	wrapperPool     sync.Pool
)

func init() {
	structFieldPool.New = func() interface{} {
		s := make([]structField, 0, 16)
		return &s
	}
	wrapperPool.New = func() interface{} {
		w := make(map[string]*json.RawMessage, 32)
		return &w
	}
}

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
func Unmarshal(msg []byte, v Attributes) error {
	val := reflect.Indirect(reflect.ValueOf(v))
	if !val.IsValid() {
		return errors.New("Unmarshal called with nil AttrSetter")
	}
	wrapperp := wrapperPool.Get().(*map[string]*json.RawMessage)
	wrapper := *wrapperp
	defer func() {
		for k := range wrapper {
			delete(wrapper, k)
		}
		wrapperPool.Put(wrapperp)
	}()
	if err := jsoniter.Unmarshal(msg, wrapperp); err != nil {
		return err
	}

	fieldsp := structFieldPool.Get().(*[]structField)
	defer func() {
		*fieldsp = (*fieldsp)[:0]
		structFieldPool.Put(fieldsp)
	}()

	getJSONFields(val, nil, false, fieldsp)
	jsonFields := *fieldsp

	for _, field := range jsonFields {
		k := field.JSONName
		v, ok := wrapper[k]
		if !ok {
			continue
		}
		delete(wrapper, k)
		if v == nil {
			continue
		}
		if err := fastjson.Unmarshal([]byte(*v), field.Value.Addr().Interface()); err != nil {
			return err
		}
	}

	if len(wrapper) > 0 {
		attrs, err := fastjson.Marshal(wrapper)
		if err != nil {
			return err
		}
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

	fieldsp := structFieldPool.Get().(*[]structField)
	defer func() {
		*fieldsp = (*fieldsp)[:0]
		structFieldPool.Put(fieldsp)
	}()

	// populate fieldsp with structFields
	getJSONFields(strukt, extendedAttributesAddress, true, fieldsp)

	fields := *fieldsp

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

func lookupField(fields []structField, name string) (structField, bool) {
	for i := range fields {
		if fields[i].JSONName == name {
			return fields[i], true
		}
	}
	return structField{}, false
}

// getJSONFields finds all of the valid JSON fields in v. It uses the address
// of the extended attributes in order to avoid listing the extended attributes
// as a valid json field, and also has a flag to control omitEmpty behaviour.
// getJSONFields populates resulp with all of the fields it finds.
// The function uses resultp as a parameter in order to support sync pooling.
func getJSONFields(v reflect.Value, addressOfAttrs *byte, omitEmpty bool, resultp *[]structField) {
	typ := v.Type()
	numField := v.NumField()
	result := *resultp
	var sf structField
	for i := 0; i < numField; i++ {
		sf.Field = typ.Field(i)
		if len(sf.Field.PkgPath) != 0 {
			// unexported
			continue
		}
		sf.Value = v.Field(i)
		elem := reflect.Indirect(sf.Value)
		sf.JSONName, sf.OmitEmpty = sf.jsonFieldName()
		if sf.JSONName == "-" {
			continue
		}
		if omitEmpty && sf.OmitEmpty && isEmpty(sf.Value) {
			continue
		}
		if elem.IsValid() && isExtendedAttributes(addressOfAttrs, sf.Value) {
			continue
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
			fieldsp := structFieldPool.Get().(*[]structField)
			getJSONFields(sf.Value, attrAddr, omitEmpty, fieldsp)
			result = append(result, *fieldsp...)
			*fieldsp = (*fieldsp)[:0]
			structFieldPool.Put(fieldsp)
			continue
		}
		// sf is a valid JSON field.
		result = append(result, sf)
	}
	*resultp = result
}
