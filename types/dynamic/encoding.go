package dynamic

import (
	"encoding/json"
	"reflect"
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

func lookupField(fields []structField, name string) (structField, bool) {
	for i := range fields {
		if fields[i].JSONName == name {
			return fields[i], true
		}
	}
	return structField{}, false
}

// getJSONFields finds all of the valid JSON fields in v.  It has a flag to
// control omitEmpty behaviour.
// resultp is populated with all of the fields that were found.
// The function uses resultp as a parameter in order to support sync pooling.
func getJSONFields(v reflect.Value, omitEmpty bool, resultp *[]structField) {
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
		sf.JSONName, sf.OmitEmpty = sf.jsonFieldName()
		if sf.JSONName == "-" {
			continue
		}
		if omitEmpty && sf.OmitEmpty && isEmpty(sf.Value) {
			continue
		}
		// if the field is embedded, flatten it out
		if sf.Field.Anonymous {
			fieldsp := structFieldPool.Get().(*[]structField)
			getJSONFields(sf.Value, omitEmpty, fieldsp)
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
