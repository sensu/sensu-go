package v2

import (
	"fmt"
	"path"
)

const (

	// LowerBound is the minimum interval that tessen will phone home, in seconds
	LowerBound = 60

	// UpperBound is the maximum interval that tessen will phone home, in seconds
	UpperBound = 21600

	// DefaultTessenInterval is the default interval at which tessen will phone home, in seconds
	DefaultTessenInterval = 1800

	// TessenResource is the name of this resource type
	TessenResource = "tessen"
)

// GetObjectMeta only exists here to fulfil the requirements of Resource
func (t *TessenConfig) GetObjectMeta() ObjectMeta {
	return ObjectMeta{}
}

// StorePrefix returns the path prefix to the Tessen config in the store
func (t *TessenConfig) StorePrefix() string {
	return TessenResource
}

// URIPath returns the path component of the Tessen config URI.
func (t *TessenConfig) URIPath() string {
	return path.Join(URLPrefix, TessenResource)
}

// Validate validates the TessenConfig.
func (t *TessenConfig) Validate() error {
	return nil
}

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

// SetNamespace sets the namespace of the resource.
func (t *TessenConfig) SetNamespace(namespace string) {
}

// SetObjectMeta only exists here to fulfil the requirements of Resource
func (t *TessenConfig) SetObjectMeta(ObjectMeta) {
	// no-op
}

func (*TessenConfig) RBACName() string {
	return "tessen"
}
