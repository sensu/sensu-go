package types

import (
	"errors"
	fmt "fmt"
	"net/url"
)

// Validate returns an error if the environment does not pass validation tests.
func (e *Environment) Validate() error {
	if err := ValidateName(e.Name); err != nil {
		return errors.New("environment name " + err.Error())
	}

	if len(e.Organization) == 0 {
		return errors.New("environment organization must be set")
	}

	return nil
}

// FixtureEnvironment returns a mocked environment.
func FixtureEnvironment(name string) *Environment {
	return &Environment{
		Name:         name,
		Organization: "default",
	}
}

// GetEnvironment gets the Evironment that e belongs to (itself).
func (e *Environment) GetEnvironment() string {
	return e.Name
}

// Update updates an Environment with selected fields.
func (e *Environment) Update(from *Environment, fields ...string) error {
	for _, f := range fields {
		switch f {
		case "Description":
			e.Description = from.Description
		default:
			return fmt.Errorf("unsupported update field: %q", f)
		}
	}
	return nil
}

// URIPath returns the path component of a Environment URI.
func (e *Environment) URIPath() string {
	return fmt.Sprintf("/rbac/organizations/%s/environments/%s", url.PathEscape(e.Organization), url.PathEscape(e.Name))
}
