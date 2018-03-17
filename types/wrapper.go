package types

import (
	"encoding/json"
	"fmt"
)

// Wrapper is a generic wrapper, with a type field for distinguishing its
// contents.
type Wrapper struct {
	// Type is the fully-qualified type name, e.g.
	// github.com/sensu/sensu-go/types.Check,
	// OR, a short-hand name that assumes a package path of
	// github.com/sensu/sensu-go/types.
	Type string `json:"type"`

	// Value is a valid Resource of concrete type Type.
	Value Resource `json:"spec"`
}

type rawWrapper struct {
	Type  string           `json:"type"`
	Value *json.RawMessage `json:"spec"`
}

// UnmarshalJSON ...
func (w *Wrapper) UnmarshalJSON(b []byte) error {
	var wrapper rawWrapper
	if err := json.Unmarshal(b, &wrapper); err != nil {
		return fmt.Errorf("error parsing spec: %s", err)
	}
	w.Type = wrapper.Type
	resource, err := ResolveResource(wrapper.Type)
	if err != nil {
		return fmt.Errorf("error parsing spec: %s", err)
	}
	if wrapper.Value == nil {
		return fmt.Errorf("no spec provided")
	}
	if err := json.Unmarshal(*wrapper.Value, &resource); err != nil {
		return err
	}
	w.Value = resource
	return nil
}
