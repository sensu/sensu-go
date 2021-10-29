package v2

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"

	stringsutil "github.com/sensu/sensu-go/api/core/v2/internal/stringutil"
)

const (
	// MutatorsResource is the name of this resource type
	MutatorsResource = "mutators"
)

// StorePrefix returns the path prefix to this resource in the store
func (m *Mutator) StorePrefix() string {
	return MutatorsResource
}

const (
	JavascriptMutator = "javascript"
	PipeMutator       = "pipe"
)

var validMutatorTypes = map[string]struct{}{
	JavascriptMutator: struct{}{},
	PipeMutator:       struct{}{},
}

func fmtMutatorTypes() string {
	types := make([]string, 0, len(validMutatorTypes))
	for k := range validMutatorTypes {
		types = append(types, k)
	}
	return strings.Join(types, ", ")
}

// URIPath returns the path component of a mutator URI.
func (m *Mutator) URIPath() string {
	if m.Namespace == "" {
		return path.Join(URLPrefix, MutatorsResource, url.PathEscape(m.Name))
	}
	return path.Join(URLPrefix, "namespaces", url.PathEscape(m.Namespace), MutatorsResource, url.PathEscape(m.Name))
}

// Validate returns an error if the mutator does not pass validation tests.
func (m *Mutator) Validate() error {
	if err := ValidateName(m.Name); err != nil {
		return errors.New("mutator name " + err.Error())
	}
	if m.Command == "" && m.Eval == "" {
		return errors.New("mutator command or eval must be set")
	}
	if m.Type == JavascriptMutator && m.Command != "" {
		return errors.New(`"command" used with javascript mutator, should be "eval"`)
	}

	if m.Namespace == "" {
		return errors.New("namespace must be set")
	}

	if m.Type != "" {
		if _, ok := validMutatorTypes[m.Type]; !ok {
			return fmt.Errorf("invalid mutator type %q, valid types are %q", m.Type, fmtMutatorTypes())
		}
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

//
// Sorting

type cmpMutator func(a, b *Mutator) bool

// SortMutatorsByPredicate is used to sort a given collection using a given predicate.
func SortMutatorsByPredicate(hs []*Mutator, fn cmpMutator) sort.Interface {
	return &mutatorSorter{mutators: hs, byFn: fn}
}

// SortMutatorsByName is used to sort a given collection of mutators by their names.
func SortMutatorsByName(hs []*Mutator, asc bool) sort.Interface {
	if asc {
		return SortMutatorsByPredicate(hs, func(a, b *Mutator) bool {
			return a.Name < b.Name
		})
	}

	return SortMutatorsByPredicate(hs, func(a, b *Mutator) bool {
		return a.Name > b.Name
	})
}

type mutatorSorter struct {
	mutators []*Mutator
	byFn     cmpMutator
}

// Len implements sort.Interface
func (s *mutatorSorter) Len() int {
	return len(s.mutators)
}

// Swap implements sort.Interface
func (s *mutatorSorter) Swap(i, j int) {
	s.mutators[i], s.mutators[j] = s.mutators[j], s.mutators[i]
}

// Less implements sort.Interface
func (s *mutatorSorter) Less(i, j int) bool {
	return s.byFn(s.mutators[i], s.mutators[j])
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
	fields := map[string]string{
		"mutator.name":           resource.ObjectMeta.Name,
		"mutator.namespace":      resource.ObjectMeta.Namespace,
		"mutator.runtime_assets": strings.Join(resource.RuntimeAssets, ","),
	}
	stringsutil.MergeMapWithPrefix(fields, resource.ObjectMeta.Labels, "mutator.labels.")
	return fields
}

// SetNamespace sets the namespace of the resource.
func (m *Mutator) SetNamespace(namespace string) {
	m.Namespace = namespace
}

// SetObjectMeta sets the meta of the resource.
func (m *Mutator) SetObjectMeta(meta ObjectMeta) {
	m.ObjectMeta = meta
}

func (m *Mutator) RBACName() string {
	return "mutators"
}
