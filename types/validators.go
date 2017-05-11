package types

import (
	"errors"
	"regexp"
)

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

// validateName validates the name of an element so it's not empty and it does
// not contains specical characters
func validateName(name string) error {
	if name == "" {
		return errors.New("must not be empty")
	}

	r, _ := regexp.Compile(`\A[\w.-]+\z`)
	match := r.MatchString(name)
	if !match {
		return errors.New("cannot contain spaces or special characters")
	}

	return nil
}
