package types

import (
	"errors"
	"fmt"
	"net/url"
)

// Validate validates the extension.
func (e *Extension) Validate() error {
	if err := ValidateName(e.Name); err != nil {
		return err
	}
	if e.URL == "" {
		return errors.New("empty URL")
	}
	if e.Namespace == "" {
		return errors.New("empty namespace")
	}
	return nil
}

// URIPath returns the path component of an Extension URI.
func (e *Extension) URIPath() string {
	return fmt.Sprintf("/extensions/%s", url.PathEscape(e.Name))
}

// FixtureExtension given a name returns a valid extension for use in tests
func FixtureExtension(name string) *Extension {
	return &Extension{
		URL: "https://localhost:8080",
		ObjectMeta: ObjectMeta{
			Namespace: "default",
			Name:      name,
		},
	}
}
