package types

import "errors"

// Validate validates the extension.
func (e *Extension) Validate() error {
	if err := ValidateName(e.Name); err != nil {
		return err
	}
	if e.URL == "" {
		return errors.New("empty URL")
	}
	if e.Organization == "" {
		return errors.New("empty Organization")
	}
	return nil
}

// FixtureExtension given a name returns a valid extension for use in tests
func FixtureExtension(name string) *Extension {
	return &Extension{
		Name:         name,
		URL:          "https://localhost:8080",
		Organization: "default",
	}
}

// GetEnvironment gets the environment for the extension. For extensions,
// the environment is always '*'.
func (e *Extension) GetEnvironment() string {
	return "*"
}
