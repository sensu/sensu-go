package dynamic

import (
	"reflect"

	utilstrings "github.com/sensu/sensu-go/util/strings"
)

// Redacted represents the redacted value for string fields
const Redacted = "REDACTED"

// DefaultRedactFields contains the default fields to redact
var DefaultRedactFields = []string{"password", "passwd", "pass", "api_key",
	"api_token", "access_key", "secret_key", "private_key", "secret"}

// Redact recursively loops through v in order to redact sensitive fields
// specified in fields and returns the redacted version of v
func Redact(v interface{}, fields ...string) (interface{}, error) {
	original := reflect.Indirect(reflect.ValueOf(v))
	redacted := reflect.New(original.Type()).Elem()

	// Use the default fields to redact in case we received none
	if len(fields) == 0 {
		fields = DefaultRedactFields
	}

	if err := redactValue(original, redacted, "", fields); err != nil {
		return nil, err
	}

	return redacted.Addr().Interface(), nil
}

// redactValue is a recursive function that redacts the original value, into
// redacted value, based on its type
func redactValue(original, redacted reflect.Value, fieldName string, fields []string) error {
	// Verify if this field is configured to be redacted
	if utilstrings.FoundInArray(fieldName, fields) {
		if original.Kind() == reflect.Interface {
			if reflect.TypeOf(original.Elem().Interface()).Kind() == reflect.String {
				redacted.Set(reflect.ValueOf(Redacted))
				return nil
			}
		}
		if original.Kind() == reflect.String {
			redacted.SetString(Redacted)
			return nil
		}

		// Set the value to the type's default
		redacted.Set(reflect.Zero(original.Type()))

		return nil
	}

	switch original.Kind() {
	case reflect.Interface:
		if err := redactInterface(original, redacted, fieldName, fields); err != nil {
			return err
		}
	case reflect.Map:
		if err := redactMap(original, redacted, fieldName, fields); err != nil {
			return err
		}
	case reflect.Ptr:
		if err := redactPtr(original, redacted, fieldName, fields); err != nil {
			return err
		}
	case reflect.Struct:
		if err := redactStruct(original, redacted, fieldName, fields); err != nil {
			return err
		}
	default:
		// Not redacted, set the original value
		redacted.Set(original)
	}

	return nil
}

// redactInterface retrieves the actual value and type of original interface and
// calls back redactValue with the right value
func redactInterface(original, redacted reflect.Value, fieldName string, fields []string) error {
	// Get actual value of original by getting rid of the interface
	originalValue := original.Elem()

	// Make sure it's not empty
	if isEmpty(originalValue) {
		return nil
	}

	// Initialize a new object with the original type
	redactedValue := reflect.New(originalValue.Type()).Elem()

	// Redact values within the interface value
	if err := redactValue(originalValue, redactedValue, fieldName, fields); err != nil {
		return err
	}

	redacted.Set(redactedValue)

	return nil
}

// redactMap loops through every key of the original value map in order to
// redact any of these keys found in fields
func redactMap(original, redacted reflect.Value, fieldName string, fields []string) error {
	// Make sure it's not empty
	if isEmpty(original) {
		return nil
	}

	// Initialize a map into redacted so we can assign keys to it
	redacted.Set(reflect.MakeMap(original.Type()))

	// Loop through the keys of the original map
	for _, key := range original.MapKeys() {
		// Get the value associated with key in the map original
		originalValue := original.MapIndex(key)

		// Initialize a new object with the type of the key's value
		redactedValue := reflect.New(originalValue.Type()).Elem()

		// Redact values in that key
		if err := redactValue(originalValue, redactedValue, key.String(), fields); err != nil {
			return err
		}

		redacted.SetMapIndex(key, redactedValue)
	}

	return nil
}

func redactPtr(original, redacted reflect.Value, fieldName string, fields []string) error {
	// Get actual value of original by getting rid of the pointer
	originalValue := original.Elem()

	// Make sure it's not nil
	if !originalValue.IsValid() {
		return nil
	}

	// Initialize a variable in redacted using the pointer value type
	redacted.Set(reflect.New(originalValue.Type()))

	// Redact values within the pointer value
	return redactValue(originalValue, redacted.Elem(), fieldName, fields)
}

// redactStruct loops through every field of the original value struct in order
// to redact any of these fields found in fields
func redactStruct(original, redacted reflect.Value, fieldName string, fields []string) error {
	// Make sure it's not empty
	if isEmpty(original) {
		return nil
	}

	// Loop through every field and call back redactValue
	for i := 0; i < original.NumField(); i++ {
		if err := redactValue(original.Field(i), redacted.Field(i), original.Type().Field(i).Name, fields); err != nil {
			return err
		}
	}

	return nil
}
