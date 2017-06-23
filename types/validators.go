package types

import (
	"errors"
	"regexp"
)

// NameRegex is used to validate the name of a resource
var NameRegex = regexp.MustCompile(`\A[\w.-]+\z`)

// StrictNameRegex is used to validate names of resources using a strict subset of charset.
var StrictNameRegex = regexp.MustCompile(`^[a-z0-9\_\.\-]+$`)

func validateHandlerType(t string) error {
	if t == "" {
		return errors.New("must not be empty")
	}

	switch t {
	case
		"pipe",
		"tcp",
		"udp",
		"transport",
		"set":
		return nil
	}

	return errors.New("is unknown")
}

// ValidateName validates the name of an element so it's not empty and it does
// not contains specical characters. Compatible with Sensu 1.0.
func ValidateName(name string) error {
	return validateNameWithPattern(name, NameRegex)
}

// ValidateNameStrict validates the name of an element so it's not empty and it
// does not contains specical characters. Compatible with Sensu 1.0.
func ValidateNameStrict(name string) error {
	return validateNameWithPattern(name, StrictNameRegex)
}

// validateName validates the name of an element so it's not empty and it does
// not contains specical characters
func validateNameWithPattern(name string, rexp *regexp.Regexp) error {
	if name == "" {
		return errors.New("must not be empty")
	}

	if match := rexp.MatchString(name); !match {
		return errors.New("cannot contain spaces or special characters")
	}

	return nil
}
