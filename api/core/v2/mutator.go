package v2

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

const (
	// MutatorsResource is the name of this resource type
	MutatorsResource = "mutators"
)

// StorePath returns the path prefix to mutators in the store
func (m *Mutator) StorePath() string {
	return MutatorsResource
}

// URIPath returns the path component of a mutator URI.
func (m *Mutator) URIPath() string {
	return path.Join(URLPrefix, "namespaces", url.PathEscape(m.Namespace), MutatorsResource, url.PathEscape(m.Name))
}

// Validate returns an error if the mutator does not pass validation tests.
func (m *Mutator) Validate() error {
	if err := ValidateName(m.Name); err != nil {
		return errors.New("mutator name " + err.Error())
	}
	if m.Command == "" {
		return errors.New("mutator command must be set")
	}

	if m.Namespace == "" {
		return errors.New("namespace must be set")
	}

	return nil
}

// Update updates m with selected fields. Returns non-nil error if any of the
// selected fields are unsupported.
func (m *Mutator) Update(from *Mutator, fields ...string) error {
	for _, f := range fields {
		switch f {
		case "Command":
			m.Command = from.Command
		case "Timeout":
			m.Timeout = from.Timeout
		case "EnvVars":
			m.EnvVars = append(m.EnvVars[0:0], from.EnvVars...)
		case "RuntimeAssets":
			m.RuntimeAssets = append(m.RuntimeAssets[0:0], from.RuntimeAssets...)
		default:
			return fmt.Errorf("unsupported field: %q", f)
		}
	}
	return nil
}

// NewMutator creates a new Mutator.
func NewMutator(meta ObjectMeta) *Mutator {
	return &Mutator{ObjectMeta: meta}
}

// FixtureMutator returns a Mutator fixture for testing.
func FixtureMutator(name string) *Mutator {
	return &Mutator{
		Command:    "command",
		ObjectMeta: NewObjectMeta(name, "default"),
	}
}

// MutatorFields returns a set of fields that represent that resource
func MutatorFields(r Resource) map[string]string {
	resource := r.(*Mutator)
	return map[string]string{
		"mutator.name":           resource.ObjectMeta.Name,
		"mutator.namespace":      resource.ObjectMeta.Namespace,
		"mutator.runtime_assets": strings.Join(resource.RuntimeAssets, ","),
	}
}
