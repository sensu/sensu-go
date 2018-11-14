package dynamic

import (
	"reflect"
	"strings"
	"sync"
)

var jsonFieldCache = new(sync.Map)

// structField is an internal convenience type
type structField struct {
	Field     reflect.StructField
	Value     reflect.Value
	JSONName  string
	OmitEmpty bool
}

type jsonField struct {
	name      string
	omitEmpty bool
}

func (s *structField) jsonFieldName() (string, bool) {
	cached, ok := jsonFieldCache.Load(s.Field.Tag)
	if ok {
		field := cached.(jsonField)
		return field.name, field.omitEmpty
	}
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
	jsonFieldCache.Store(s.Field.Tag, jsonField{name: fieldName, omitEmpty: omitEmpty})
	return fieldName, omitEmpty
}
