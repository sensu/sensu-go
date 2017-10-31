package types

import "errors"

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

// GetOrg refers to the organization the check belongs to
func (m *Mutator) GetOrg() string {
	return m.Organization
}

// GetEnv refers to the organization the check belongs to
func (m *Mutator) GetEnv() string {
	return m.Environment
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
