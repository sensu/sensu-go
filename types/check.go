package types

import (
	"errors"
	"sort"
	"time"
)

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

	// Handlers are the event handler for the check (incidents
	// and/or metrics).
	Handlers []string `json:"handlers"`

	// Output from the execution of Command.
	Output string `json:"output,omitempty"`

	// Status is the exit status code produced by the check.
	Status int `json:"status,omitempty"`

	// Time check request was issued.
	Issued int64 `json:"issued,omitempty"`

	// Time check request was executed
	Executed int64 `json:"executed,omitempty"`

	// Duration of execution.
	Duration float64 `json:"duration,omitempty"`

	// History is the check state history.
	History []CheckHistory `json:"history,omitempty"`
}

// Validate returns an error if the check does not pass validation tests.
func (c *Check) Validate() error {
	err := validateName(c.Name)
	if err != nil {
		return errors.New("check name " + err.Error())
	}

	if c.Interval <= 0 {
		return errors.New("check interval must be greater than 0")
	}

	if c.Command == "" {
		return errors.New("check command must be set")
	}

	return nil
}

// CheckHistory is a record of a check execution and its status.
type CheckHistory struct {
	Status   int   `json:"status"`
	Executed int64 `json:"executed"`
}

// ByExecuted implements the sort.Interface for []CheckHistory based on the
// Executed field.
//
// Example:
//
// sort.Sort(ByExecuted(check.History))
type ByExecuted []CheckHistory

func (b ByExecuted) Len() int           { return len(b) }
func (b ByExecuted) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByExecuted) Less(i, j int) bool { return b[i].Executed < b[j].Executed }

// MergeWith updates the current Check with the history of the check given as
// an argument, updating the current check's history appropriately.
func (c *Check) MergeWith(chk *Check) {
	history := chk.History
	histEntry := CheckHistory{
		Status:   chk.Status,
		Executed: chk.Executed,
	}

	history = append([]CheckHistory{histEntry}, history...)
	sort.Sort(ByExecuted(history))
	if len(history) > 21 {
		history = history[1:]
	}

	c.History = history
}

// FixtureCheck returns a fixture for a Check object.
func FixtureCheck(id string) *Check {
	t := time.Now().Unix()
	interval := 60

	history := make([]CheckHistory, 21)
	for i := 0; i < 21; i++ {
		history[i] = CheckHistory{
			Status:   0,
			Executed: t - (60 * int64(i+1)),
		}
	}

	return &Check{
		Name:          id,
		Interval:      interval,
		Subscriptions: []string{},
		Command:       "command",
		Status:        0,
		Issued:        t,
		Executed:      t + 1,
		Duration:      1.0,
		Handlers:      []string{},
		History:       history,
	}
}
