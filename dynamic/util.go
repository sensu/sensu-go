package dynamic

import (
	"reflect"
	"strings"
)

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
