package conversion

import (
	"errors"
	"fmt"
)

// ErrImpossibleConversion indicates that an incompatible type conversion was
// attempted.
var ErrImpossibleConversion = errors.New("impossible type conversion")

type key struct {
	SourceAPIVersion string
	DestAPIVersion   string
	Kind             string
}

// Interface specifies the requirements for the ConvertTypes function.
type Interface interface {
	GetKind() string
	GetAPIVersion() string
}

// registry will be populated by init functions that will be generated in this
// package.
var registry = map[key]conversionFunc{}

type conversionFunc func(dst, src interface{}) error

// ConvertTypes converts a source type to a destination type.
// The dst parameter must be addressable, and will be written to if a valid
// conversion exists.
//
// ConvertTypes relies on an internal registry which is automatically generated
// from the internal and versioned API types. It is not possible to convert
// between versioned types; instead, users must convert to internal types, and
// then convert the internal type to a versioned type.
//
// If the types passed to ConvertTypes are incompatible, that is, no conversion
// method exists, ErrImpossibleConversion will be returned.
func ConvertTypes(dst, src Interface) (err error) {
	key := key{
		SourceAPIVersion: src.GetAPIVersion(),
		DestAPIVersion:   dst.GetAPIVersion(),
		Kind:             dst.GetKind(),
	}
	fn, ok := registry[key]
	if !ok {
		return ErrImpossibleConversion
	}
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("error while performing conversion: %v", e)
		}
	}()
	return fn(dst, src)
}
