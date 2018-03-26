package types

import (
	"errors"
	fmt "fmt"
	"net/url"
)

// Validate returns an error if the mutator does not pass validation tests.
func (m *Mutator) Validate() error {
	if err := ValidateName(m.Name); err != nil {
		return errors.New("mutator name " + err.Error())
	}
	if m.Command == "" {
		return errors.New("mutator command must be set")
	}

	if m.Environment == "" {
		return errors.New("mutator environment must be set")
	}

	if m.Organization == "" {
		return errors.New("mutator organization must be set")
	}

	return nil
}

// Update updates m with selected fields. Returns non-nil error if any of the
// selected fields are unsupported.
func (m *Mutator) Update(from *Mutator, fields ...string) error {
	for _, f := range fields {
		switch f {
		case "Command":
			m.Command = from.Command
		case "Timeout":
			m.Timeout = from.Timeout
		case "EnvVars":
			m.EnvVars = append(m.EnvVars[0:0], from.EnvVars...)
		default:
			return fmt.Errorf("unsupported field: %q", f)
		}
	}
	return nil
}

// FixtureMutator returns a Mutator fixture for testing.
func FixtureMutator(name string) *Mutator {
	return &Mutator{
		Name:         name,
		Command:      "command",
		Environment:  "default",
		Organization: "default",
	}
}

// URIPath returns the path component of a Mutator URI.
func (m *Mutator) URIPath() string {
	return fmt.Sprintf("/mutators/%s", url.PathEscape(m.Name))
}
