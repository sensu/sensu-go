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

type lifter interface {
	Lift() Resource
}

// UnmarshalJSON implements json.Unmarshaler
func (w *Wrapper) UnmarshalJSON(b []byte) error {
	// Decode the top-level fields only of the incoming data into a temporary
	// rawWrapper variable
	var wrapper rawWrapper
	if err := json.Unmarshal(b, &wrapper); err != nil {
		return fmt.Errorf("error parsing data as wrapped-json: %s", err)
	}

	// Set the TypeMeta on the wrapper
	w.TypeMeta = wrapper.TypeMeta
	if w.APIVersion == "" {
		w.APIVersion = "core/v2"
	}

	// Use the TypeMeta to resolve the type of the resource contained in the Value
	// field as a *json.RawMessage
	resource, err := ResolveType(w.TypeMeta.APIVersion, w.TypeMeta.Type)
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

	// Special case for the Namespace resource
	if _, ok := resource.(*Namespace); ok {
		w.Value = resource
		return nil
	}

	// Use the outer ObjectMeta to fill the inner ObjectMeta that's part of the
	// resource if it's empty
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

	// Set the outer ObjectMeta of the wrapper
	w.ObjectMeta = outerMeta

	// Set the inner ObjectMeta
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

	// Determine if the resource implements the Lifter interface, which has a Lift
	// method. This is useful when a resource can be polymorphic, such as
	// providers.
	if lifter, ok := resource.(lifter); ok {
		resource = lifter.Lift()
	}

	// Set the resource as the wrapper's value
	w.Value = resource

	return nil
}

// tmGetter is useful for types that want to explicitly provide their
// TypeMeta - this is useful for lifters.
type tmGetter interface {
	GetTypeMeta() TypeMeta
}

// WrapResource wraps a Resource in a Wrapper that contains TypeMeta and
// ObjectMeta.
func WrapResource(r Resource) Wrapper {
	var tm TypeMeta
	if getter, ok := r.(tmGetter); ok {
		tm = getter.GetTypeMeta()
	} else {
		typ := reflect.Indirect(reflect.ValueOf(r)).Type()
		tm = TypeMeta{
			Type:       typ.Name(),
			APIVersion: ApiVersion(typ.PkgPath()),
		}
	}
	return Wrapper{
		TypeMeta:   tm,
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

// ResolveType returns the Resource associated with the given package and type.
func ResolveType(apiVersion string, typename string) (Resource, error) {
	// Guard read access to packageMap
	packageMapMu.RLock()
	defer packageMapMu.RUnlock()
	resolver, ok := packageMap[apiVersion]
	if !ok {
		return nil, fmt.Errorf("invalid API version: %s", apiVersion)
	}
	return resolver(typename)
}

func ApiVersion(version string) string {
	parts := strings.Split(version, "/")
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return path.Join(parts[len(parts)-2], parts[len(parts)-1])
}
