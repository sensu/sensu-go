package v2

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
	return fmt.Sprintf("/api/core/v2/namespaces/%s/extensions/%s", url.PathEscape(e.Namespace), url.PathEscape(e.Name))
}

// FixtureExtension given a name returns a valid extension for use in tests
func FixtureExtension(name string) *Extension {
	return &Extension{
		URL:        "https://localhost:8080",
		ObjectMeta: NewObjectMeta(name, "default"),
	}
}

// NewExtension intializes an extension with the given object meta
func NewExtension(meta ObjectMeta) *Extension {
	return &Extension{ObjectMeta: meta}
}

// ExtensionFields returns a set of fields that represent that resource
func ExtensionFields(r Resource) map[string]string {
	resource := r.(*Extension)
	return map[string]string{
		"extension.name":      resource.ObjectMeta.Name,
		"extension.namespace": resource.ObjectMeta.Namespace,
	}
}
