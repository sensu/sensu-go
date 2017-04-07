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
}

// Validate returns an error if the mutator does not pass validation tests.
func (c *Mutator) Validate() error {
	if c.Name == "" {
		return errors.New("name cannot be empty")
	}

	if c.Command == "" {
		return errors.New("must have a command")
	}

	return nil
}
