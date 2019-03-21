package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"reflect"
	"strings"
	"sync"

	v2 "github.com/sensu/sensu-go/api/core/v2"
)

// Wrapper is a generic wrapper, with a type field for distinguishing its
// contents.
type Wrapper struct {
	TypeMeta

	ObjectMeta ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Value is a valid Resource of concrete type Type.
	Value Resource `json:"spec" yaml:"spec"`
}

type rawWrapper struct {
	TypeMeta
	ObjectMeta ObjectMeta       `json:"metadata" yaml:"metadata"`
	Value      *json.RawMessage `json:"spec" yaml:"spec"`
}

// PackageMap contains a list of packages with their Resource Resolver func
var packageMap = map[string]func(string) (Resource, error){
	"core/v2": v2.ResolveResource,
}

var packageMapMu = &sync.RWMutex{}

// UnmarshalJSON implements json.Unmarshaler
func (w *Wrapper) UnmarshalJSON(b []byte) error {
	var wrapper rawWrapper
	if err := json.Unmarshal(b, &wrapper); err != nil {
		return fmt.Errorf("error parsing spec: %s", err)
	}
	w.TypeMeta = wrapper.TypeMeta
	if w.APIVersion == "" {
		w.APIVersion = "core/v2"
	}

	// Guard read access to packageMap
	packageMapMu.RLock()
	defer packageMapMu.RUnlock()
	resolver, ok := packageMap[w.TypeMeta.APIVersion]
	if !ok {
		return fmt.Errorf("invalid API version: %s", w.TypeMeta.APIVersion)
	}
	resource, err := resolver(w.TypeMeta.Type)
	if err != nil {
		return fmt.Errorf("error parsing spec: %s", err)
	}
	if wrapper.Value == nil {
		return fmt.Errorf("no spec provided")
	}
	dec := json.NewDecoder(bytes.NewReader(*wrapper.Value))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&resource); err != nil {
		return err
	}
	if _, ok := resource.(*Namespace); ok {
		// Special case for Namespace
		w.Value = resource
		return nil
	}
	outerMeta := wrapper.ObjectMeta
	innerMeta := resource.GetObjectMeta()
	if outerMeta.Namespace != "" {
		innerMeta.Namespace = outerMeta.Namespace
	}
	if outerMeta.Name != "" {
		innerMeta.Name = outerMeta.Name
	}
	for k, v := range outerMeta.Labels {
		if innerMeta.Labels == nil {
			innerMeta.Labels = make(map[string]string)
		}
		innerMeta.Labels[k] = v
	}
	for k, v := range outerMeta.Annotations {
		if innerMeta.Annotations == nil {
			innerMeta.Annotations = make(map[string]string)
		}
		innerMeta.Annotations[k] = v
	}
	val := reflect.Indirect(reflect.ValueOf(resource))
	objectMeta := val.FieldByName("ObjectMeta")
	if objectMeta.Kind() == reflect.Invalid {
		// The resource doesn't have an ObjectMeta field - this is expected
		// for Namespace, or other types that have no ObjectMeta field but
		// do implement a GetObjectMeta method.
		w.Value = resource
		return nil
	}
	val.FieldByName("ObjectMeta").Set(reflect.ValueOf(innerMeta))
	w.Value = resource
	return nil
}

// WrapResource wraps a Resource in a Wrapper that contains TypeMeta and
// ObjectMeta.
func WrapResource(r Resource) Wrapper {
	typ := reflect.Indirect(reflect.ValueOf(r)).Type()
	return Wrapper{
		TypeMeta: TypeMeta{
			Type:       typ.Name(),
			APIVersion: apiVersion(typ.PkgPath()),
		},
		ObjectMeta: r.GetObjectMeta(),
		Value:      r,
	}
}

// RegisterTypeResolver adds a package to packageMap with its resolver
func RegisterTypeResolver(key string, resolver func(string) (Resource, error)) {
	packageMapMu.Lock()
	defer packageMapMu.Unlock()
	packageMap[key] = resolver
}

func apiVersion(version string) string {
	parts := strings.Split(version, "/")
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return path.Join(parts[len(parts)-2], parts[len(parts)-1])
}
