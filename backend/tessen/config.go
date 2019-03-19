package tessen

import (
	"fmt"
)

const (
	// LowerBound is the minimum frequency that tessen will phone home
	LowerBound = 1

	// UpperBound is the maximum frequency that tessen will phone home
	UpperBound = 360

	// DefaultTessenFrequency is the default frequency at which tessen will phone home
	DefaultTessenFrequency = 1
)

// Config is the tessen configuration
type Config struct {
	OptOut bool `json:"opt_out"`
}

// ValidateFrequency returns an error if the tessen frequency is not within the upper and lower bound limits
func ValidateFrequency(freq uint32) error {
	if !(freq >= LowerBound && freq <= UpperBound) {
		return fmt.Errorf("tessen frequency must be in between the lower and upper bound limits (%d-%d)", LowerBound, UpperBound)
	}
	return nil
}

// DefaultConfig returns the default tessen configuration
func DefaultConfig() *Config {
	return &Config{}
}
