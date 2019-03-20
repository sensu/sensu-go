package v2

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

	// TessenPath is the store and api path for tessen
	TessenPath = "/tessen"
)

// ValidateFrequency returns an error if the tessen frequency is not within the upper and lower bound limits
func ValidateFrequency(freq uint32) error {
	if !(freq >= LowerBound && freq <= UpperBound) {
		return fmt.Errorf("tessen frequency must be in between the lower and upper bound limits (%d-%d)", LowerBound, UpperBound)
	}
	return nil
}

// DefaultTessenConfig returns the default tessen configuration
func DefaultTessenConfig() *TessenConfig {
	return &TessenConfig{}
}

// URIPath returns the path component of a TessenConfig URI.
func (t *TessenConfig) URIPath() string {
	return fmt.Sprintf("/api/core/v2%s", TessenPath)
}

// Validate validates the TessenConfig.
func (t *TessenConfig) Validate() error {
	return nil
}

// GetObjectMeta only exists here to fulfil the requirements of Resource
func (t *TessenConfig) GetObjectMeta() ObjectMeta {
	return ObjectMeta{}
}
