package types

import "errors"

// A Check is a check specification and optionally the results of the check's
// execution.
type Check struct {
	// Name is the unique identifier for a check.
	Name string `json:"name"`

	// Interval is the interval, in seconds, at which the check should be
	// run.
	Interval int `json:"interval"`

	// Subscriptions is the list of subscribers for the check.
	Subscriptions []string `json:"subscriptions"`

	// Command is the command to be executed.
	Command string `json:"command"`

	// Output from the execution of Command.
	Output string `json:"output"`

	// Status from the execution of Command.
	Status int64 `json:"output"`

	// Time check request was issued.
	Issued int64 `json:"issued"`

	// Time check request was executed
	Executed int64 `json:"executed"`

	// Duration of execution.
	Duration float64 `json:"duration"`

	// Handlers are the event handler for the check (incidents
	// and/or metrics).
	Handlers []string `json:"handlers"`
}

// Validate returns an error if the check does not pass validation tests.
func (c *Check) Validate() error {
	if c.Name == "" {
		return errors.New("name cannot be empty")
	}

	if c.Interval == 0 {
		return errors.New("interval must be greater than zero")
	}

	if c.Command == "" {
		return errors.New("must have a command")
	}

	return nil
}
