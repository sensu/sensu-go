package v3

// automatically generated file, do not edit!

import (
	"fmt"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"reflect"
)

// typeMap is used to dynamically look up data types from strings.
var typeMap = map[string]interface{}{
	"EntityConfig":  &EntityConfig{},
	"entity_config": &EntityConfig{},
	"EntityState":   &EntityState{},
	"entity_state":  &EntityState{},
}

// ResolveResource returns a zero-valued resource, given a name.
// If the named type does not exist, or if the type is not a Resource,
// then an error will be returned.
func ResolveResource(name string) (corev2.Resource, error) {
	t, ok := typeMap[name]
	if !ok {
		return nil, fmt.Errorf("type could not be found: %q", name)
	}
	if _, ok := t.(corev2.Resource); !ok {
		return nil, fmt.Errorf("%q is not a Resource", name)
	}
	return newResource(t), nil
}

// Make a new Resource to avoid aliasing problems with ResolveResource.
// don't use this function. no, seriously.
func newResource(r interface{}) corev2.Resource {
	return reflect.New(reflect.ValueOf(r).Elem().Type()).Interface().(corev2.Resource)
}
