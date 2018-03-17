package create

import (
	"fmt"
	"reflect"
)

// make a best-effort attempt to set org and env defaults.
// if it can't be done, return silently.
//
// TODO: handle this sort of thing in the HTTP API, with
// URL parameters or something.
func reflectSetField(v interface{}, fieldName, value string) {
	strukt := reflect.Indirect(reflect.ValueOf(v))
	if strukt.Kind() != reflect.Struct {
		// The type is not a struct, do nothing.
		return
	}
	field := strukt.FieldByName(fieldName)
	if !field.IsValid() {
		// The field doesn't exist, do nothing.
		return
	}
	fieldV := reflect.Indirect(field)
	if !fieldV.CanSet() || fieldV.Kind() != reflect.String {
		// egregious programmer error, deserving a panic.
		panic(fmt.Sprintf("unexpected datatype: %#v", v))
	}
	fieldV.SetString(value)
}
