package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"reflect"
	"strings"
	"sync"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

var ErrAPINotFound = errors.New("api not found")

// compatibility shim - can't import core/v3
type corev3Resource interface {
	GetMetadata() *corev2.ObjectMeta
	SetMetadata(*corev2.ObjectMeta)
}

// getObjectMeta gets the object metadata from either a corev2 or corev3 resource.
// It panics if the value passed is not a corev2 or corev3 resource.
func getObjectMeta(value interface{}) *corev2.ObjectMeta {
	switch value := value.(type) {
	case corev2.Resource:
		meta := value.GetObjectMeta()
		return &meta
	case corev3Resource:
		return value.GetMetadata()
	}
	// impossible unless the type resolver is broken. fatal error.
	panic("got neither corev2 resource nor corev3 resource")
}

func setObjectMeta(value interface{}, meta *corev2.ObjectMeta) {
	switch value := value.(type) {
	case corev2.Resource:
		value.SetObjectMeta(*meta)
	case corev3Resource:
		value.SetMetadata(meta)
	default:
		// impossible unless the type resolver is broken. fatal error.
		panic("got neither corev2 resource nor corev3 resource")
	}
}

// Wrapper is a generic wrapper, with a type field for distinguishing its
// contents.
type Wrapper struct {
	TypeMeta

	ObjectMeta ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// Value is a valid Resource of concrete type Type.
	Value interface{} `json:"spec" yaml:"spec"`
}

type rawWrapper struct {
	TypeMeta
	ObjectMeta ObjectMeta       `json:"metadata" yaml:"metadata"`
	Value      *json.RawMessage `json:"spec" yaml:"spec"`
}

// PackageMap contains a list of packages with their Resource Resolver func
var packageMap = map[string]interface{}{
	"core/v2": corev2.ResolveResource,
}

var packageMapMu = &sync.RWMutex{}

type lifter interface {
	Lift() Resource
}

// toMap produces a map from a struct by serializing it to JSON and then
// deserializing the JSON into a map. This is done to preserve business logic
// expressed in customer marshalers, and JSON struct tag semantics.
func toMap(v interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	err = dec.Decode(&result)
	return result, err
}

// MarshalJSON implements json.Marshaler
func (w Wrapper) MarshalJSON() ([]byte, error) {
	wrapper := struct {
		TypeMeta
		ObjectMeta ObjectMeta             `json:"metadata"`
		Value      map[string]interface{} `json:"spec"`
	}{
		TypeMeta:   w.TypeMeta,
		ObjectMeta: w.ObjectMeta,
	}

	// Remove the innerMeta
	value, err := toMap(w.Value)
	if err != nil {
		return nil, err
	}
	delete(value, "metadata")

	wrapper.Value = value

	return json.Marshal(wrapper)
}

// MarshalYAML implements yaml.Marshaler
func (w Wrapper) MarshalYAML() (interface{}, error) {
	wrapper := struct {
		Type       string                 `yaml:"type"`
		APIVersion string                 `yaml:"api_version"`
		ObjectMeta map[string]interface{} `yaml:"metadata"`
		Value      map[string]interface{} `yaml:"spec"`
	}{
		Type:       w.Type,
		APIVersion: w.APIVersion,
	}

	meta, err := toMap(w.ObjectMeta)
	if err != nil {
		return nil, err
	}
	wrapper.ObjectMeta = meta

	// Remove the innerMeta
	value, err := toMap(w.Value)
	if err != nil {
		return nil, err
	}
	delete(value, "metadata")
	wrapper.Value = value

	return wrapper, nil
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
	resource, err := ResolveRaw(w.TypeMeta.APIVersion, w.TypeMeta.Type)
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
	innerMeta := getObjectMeta(resource)
	if innerMeta == nil {
		innerMeta = &outerMeta
		setObjectMeta(resource, innerMeta)
		goto V3RESOURCE
	}

	// Start hacks to equalize ObjectMetas
	if outerMeta.Namespace != "" {
		innerMeta.Namespace = outerMeta.Namespace
	}
	if outerMeta.Name != "" {
		innerMeta.Name = outerMeta.Name
	}
	if outerMeta.CreatedBy != "" {
		innerMeta.CreatedBy = outerMeta.CreatedBy
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
	// End hacks to equalize ObjectMetas

V3RESOURCE:

	// Set the outer ObjectMeta of the wrapper
	w.ObjectMeta = *innerMeta

	// Set the inner ObjectMeta
	if r, ok := resource.(corev3Resource); ok {
		if innerMeta.Labels == nil {
			innerMeta.Labels = make(map[string]string)
		}
		if innerMeta.Annotations == nil {
			innerMeta.Annotations = make(map[string]string)
		}
		r.SetMetadata(innerMeta)
	} else {
		val := reflect.Indirect(reflect.ValueOf(resource))
		objectMeta := val.FieldByName("ObjectMeta")
		if objectMeta.Kind() == reflect.Invalid {
			// The resource doesn't have an ObjectMeta field - this is expected
			// for Namespace, or other types that have no ObjectMeta field but
			// do implement a GetObjectMeta method.
			w.Value = resource
			return nil
		}
		val.FieldByName("ObjectMeta").Set(reflect.Indirect(reflect.ValueOf(innerMeta)))
	}

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
		ObjectMeta: *getObjectMeta(r),
		Value:      r,
	}
}

// RegisterTypeResolver adds a package to packageMap with its resolver. Deprecated.
func RegisterTypeResolver(key string, resolver func(string) (Resource, error)) {
	packageMapMu.Lock()
	defer packageMapMu.Unlock()
	packageMap[key] = resolver
}

// RegisterResolver adds a package to packageMap with its resolver.
func RegisterResolver(key string, resolver func(string) (interface{}, error)) {
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
	switch resolver := resolver.(type) {
	case func(string) (Resource, error):
		v, err := resolver(typename)
		return v, err
	default:
		return nil, fmt.Errorf("%s does not implement v2.Resource", apiVersion)
	}
}

// ResolveRaw resolves the raw type for the requested type.
func ResolveRaw(apiVersion string, typename string) (interface{}, error) {
	// Guard read access to packageMap
	packageMapMu.RLock()
	defer packageMapMu.RUnlock()
	resolver, ok := packageMap[apiVersion]
	if !ok {
		return nil, ErrAPINotFound
	}
	switch resolver := resolver.(type) {
	case func(string) (Resource, error):
		return resolver(typename)
	case func(string) (interface{}, error):
		return resolver(typename)
	}
	return nil, fmt.Errorf("bad resolver: %T", resolver)
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
