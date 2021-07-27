package v3

// automatically generated file, do not edit!

import (
	"fmt"
	"reflect"
	"sort"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	types "github.com/sensu/sensu-go/types"
)

func init() {
	for _, v := range typeMap {
		if r, ok := v.(Resource); ok {
			rbacMap[r.RBACName()] = r
		}
	}
	for _, v := range rbacMap {
		storeMap[v.StoreName()] = v
	}
	types.RegisterResolver("core/v3", ResolveRawResource)
}

// typeMap is used to dynamically look up data types from strings.
var typeMap = map[string]interface{}{
	"EntityConfig":      &EntityConfig{},
	"entity_config":     &EntityConfig{},
	"EntityState":       &EntityState{},
	"entity_state":      &EntityState{},
	"Pipeline":          &Pipeline{},
	"pipeline":          &Pipeline{},
	"ResourceTemplate":  &ResourceTemplate{},
	"resource_template": &ResourceTemplate{},
}

// rbacMap is like typemap, but its keys are RBAC names, and its values are
// Resource values.
var rbacMap = make(map[string]Resource, len(typeMap)/2)

// storeMap is like rbacMap, but its keys are store names.
var storeMap = make(map[string]Resource, len(typeMap)/2)

// ResolveResource returns a zero-valued resource, given a name.
// If the named type does not exist, or if the type is not a Resource,
// then an error will be returned.
func ResolveResource(name string) (Resource, error) {
	t, ok := typeMap[name]
	if !ok {
		return nil, fmt.Errorf("type could not be found: %q", name)
	}
	if _, ok := t.(Resource); !ok {
		return nil, fmt.Errorf("%q is not a core/v3.Resource", name)
	}
	return newResource(t), nil
}

// ResolveRawResource is like ResolveResource, but uses interface{} instead of
// Resource as a return type.
func ResolveRawResource(name string) (interface{}, error) {
	return ResolveResource(name)
}

// Make a new Resource to avoid aliasing problems with ResolveResource.
// don't use this function. no, seriously.
func newResource(r interface{}) Resource {
	value := reflect.New(reflect.ValueOf(r).Elem().Type()).Interface().(Resource)
	value.SetMetadata(&corev2.ObjectMeta{
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	})
	return value
}

// ResolveV2Resource resolves the resources in this package to core/v2.Resource
// types, via the V2ResourceProxy type.
func ResolveV2Resource(name string) (corev2.Resource, error) {
	resource, err := ResolveResource(name)
	if err != nil {
		return nil, err
	}
	return V3ToV2Resource(resource), nil
}

// ResolveResourceByRBACName resolves a resource by its RBAC name.
func ResolveResourceByRBACName(name string) (Resource, error) {
	resource, ok := rbacMap[name]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", name)
	}
	return newResource(resource), nil
}

// ResolveResourceByStoreName resolves a resource by its store name.
func ResolveResourceByStoreName(name string) (Resource, error) {
	resource, ok := storeMap[name]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", name)
	}
	return newResource(resource), nil
}

// ListResources lists all of the resources in the package.
func ListResources() []Resource {
	result := make([]Resource, 0, len(rbacMap))
	for _, v := range rbacMap {
		result = append(result, newResource(v))
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].RBACName() < result[j].RBACName()
	})
	return result
}
