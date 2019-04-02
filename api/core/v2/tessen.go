package v2

import (
	"fmt"
)

const (
	// LowerBound is the minimum interval that tessen will phone home, in seconds
	LowerBound = 60

	// UpperBound is the maximum interval that tessen will phone home, in seconds
	UpperBound = 21600

	// DefaultTessenInterval is the default interval at which tessen will phone home, in seconds
	DefaultTessenInterval = 1800

	// TessenPath is the store and api path for tessen
	TessenPath = "/tessen"
)

// ValidateInterval returns an error if the tessen interval is not within the upper and lower bound limits
func ValidateInterval(freq uint32) error {
	if !(freq >= LowerBound && freq <= UpperBound) {
		return fmt.Errorf("tessen interval must be in between the lower and upper bound limits (%d-%d)", LowerBound, UpperBound)
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
