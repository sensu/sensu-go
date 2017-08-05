package types

import "errors"

// A Mutator is a mutator specification.
type Mutator struct {
	// Name is the unique identifier for a mutator.
	Name string `json:"name"`

	// Command is the command to be executed.
	Command string `json:"command"`

	// Timeout is the command execution timeout in seconds.
	Timeout int `json:"timeout"`

	// Env is a list of environment variables to use with command execution
	Env []string `json:"env,omitempty"`

	// Environment indicates to which env a mutator belongs to
	Environment string `json:"environment"`

	// Organization specifies the organization to which the mutator belongs.
	Organization string `json:"organization"`
}

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

// FixtureMutator returns a Mutator fixture for testing.
func FixtureMutator(name string) *Mutator {
	return &Mutator{
		Name:         name,
		Command:      "command",
		Environment:  "default",
		Organization: "default",
	}
}
