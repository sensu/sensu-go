package dynamic

import (
	"reflect"
	"sort"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

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

// isExtendedAttributes determines if the provided value correspond to the
// provided extended attributes address
func isExtendedAttributes(address *byte, value reflect.Value) bool {
	if address == nil {
		return false
	}
	if value.Kind() != reflect.Slice {
		return false
	}

	elem := reflect.Indirect(value)
	if b, ok := elem.Interface().([]byte); ok {
		if len(b) > 0 && &b[0] == address {
			return true
		}
	}

	return false
}

func isEmpty(value reflect.Value) bool {
	if !value.IsValid() {
		return true
	}

	switch value.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		if value.Len() == 0 {
			return true
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Int() == int64(0) {
			return true
		}
	case reflect.Interface, reflect.Ptr:
		if value.IsNil() {
			return true
		}
	}

	return false
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

// mapOfExtendedAttributes produces a map[string]interface{} of extended
// attributes with capitalization of the key
func mapOfExtendedAttributes(v interface{}) map[string]interface{} {
	values := reflect.ValueOf(v)
	if values.Kind() != reflect.Map {
		return nil
	}

	attrs := make(map[string]interface{})
	for _, value := range values.MapKeys() {
		if values.MapIndex(value).CanInterface() {
			typeOfValue := reflect.TypeOf(values.MapIndex(value).Interface()).Kind()
			if typeOfValue == reflect.Map || typeOfValue == reflect.Slice {
				attrs[strings.Title(value.String())] = mapOfExtendedAttributes(values.MapIndex(value).Interface())
			} else {
				if values.MapIndex(value).CanInterface() {
					attrs[strings.Title(value.String())] = values.MapIndex(value).Interface()
				}
			}
		}
	}

	return attrs
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
