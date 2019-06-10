package suggest

import (
	"errors"
	"strings"
)

var (
	separator = "/"

	ErrInvalidRef = errors.New("reference to field must take the form :group/:version/:resource/:field/:name")
)

type RefComponents struct {
	Group     string
	Name      string
	FieldPath string
}

// String prints the identifier
func (r RefComponents) String() string {
	return strings.Join([]string{r.Group, r.Name, r.FieldPath}, separator)
}

// ParseRef takes an identifier and returns it's components.
func ParseRef(ref string) (r RefComponents, err error) {
	paths := strings.SplitN(ref, separator, 4)
	if len(paths) < 3 {
		return r, ErrInvalidRef
	}
	if len(paths) > 3 {
		r.FieldPath = paths[3]
	}
	r.Group = strings.Join(paths[:2], separator)
	r.Name = paths[2]
	return
}
