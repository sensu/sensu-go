package validators

import (
	"fmt"
	"strings"
)

// BoolString is a string value that represents a bool.
type BoolString string

var (
	trueValues   = []BoolString{"t", "y", "true", "yes"}
	falseValues  = []BoolString{"f", "n", "false", "no"}
	acceptableTF = append(trueValues, falseValues...)
)

// Bool returns true if t is a truthy value. It returns false if
// t is a falsy value. It returns non-nil error if t is neither.
func (t BoolString) Bool() (bool, error) {
	for _, v := range trueValues {
		if t == v {
			return true, nil
		}
	}
	for _, v := range falseValues {
		if t == v {
			return false, nil
		}
	}
	return false, fmt.Errorf("invalid true/false value: %q", t)
}

// ValidateTrueFalse returns non-nil error if its argument is any of "t", "f",
// "true" or "false". Case insensitive.
func ValidateTrueFalse(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid value: %v", value)
	}
	bs := BoolString(strings.ToLower(s))
	_, err := bs.Bool()
	return err
}
