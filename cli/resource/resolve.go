package resource

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	corev2 "github.com/sensu/core/v2"
	apitools "github.com/sensu/sensu-api-tools"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/compat"
)

var (
	// All is all the core resource types and associated sensuctl verbs (non-namespaced resources are intentionally ordered first).
	All = []corev2.Resource{
		&corev2.Namespace{},
		&corev2.ClusterRole{},
		&corev2.ClusterRoleBinding{},
		&corev2.User{},
		&corev2.APIKey{},
		&corev2.TessenConfig{},
		&corev2.Asset{},
		&corev2.CheckConfig{},
		&corev2.Entity{},
		&corev2.Event{},
		&corev2.EventFilter{},
		&corev2.Handler{},
		&corev2.HookConfig{},
		&corev2.Mutator{},
		&corev2.Pipeline{},
		&corev2.Role{},
		&corev2.RoleBinding{},
		&corev2.Silenced{},
	}

	// synonyms provides user-friendly resource synonyms like checks, entities
	synonyms = map[string]corev2.Resource{}
)

func init() {
	for _, resource := range All {
		synonyms[resource.RBACName()] = resource
	}
}

type lifter interface {
	Lift() corev2.Resource
}

var resourceRE = regexp.MustCompile(`(\w+\/v\d+\.)?(\w+)`)

// Resolve resolves a named resource to an empty concrete type.
// The value is boxed within a corev2.Resource interface value.
func Resolve(resource string) (corev2.Resource, error) {
	if resource, ok := synonyms[resource]; ok {
		return resource, nil
	}
	matches := resourceRE.FindStringSubmatch(resource)
	if len(matches) != 3 {
		return nil, fmt.Errorf("bad resource qualifier: %s. hint: try something like core/v2.CheckConfig", resource)
	}
	apiVersion := strings.TrimSuffix(matches[1], ".")
	typeName := matches[2]
	if apiVersion == "" {
		apiVersion = "core/v2"
	}
	value, err := apitools.Resolve(apiVersion, typeName)
	if err != nil {
		return nil, err
	}
	return compat.V2Resource(value), nil
}

func dedupTypes(arg string) []string {
	typs := strings.Split(arg, ",")
	seen := make(map[string]struct{})
	result := make([]string, 0, len(typs))
	for _, t := range typs {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		if syn, ok := synonyms[t]; ok {
			w := types.WrapResource(syn)
			seen[fmt.Sprintf("%s.%s", w.APIVersion, w.Type)] = struct{}{}
		}
		result = append(result, t)
	}
	return result
}

// GetResourceRequests gets the resources based on the input.
func GetResourceRequests(actionSpec string, resources []corev2.Resource) ([]corev2.Resource, error) {
	// parse the comma separated resource types and match against the defined actions
	if actionSpec == "all" {
		return resources, nil
	}
	if actionSpec == "" {
		// There were no specs, return an empty slice
		return []corev2.Resource{}, nil
	}
	var actions []corev2.Resource
	// deduplicate requested resources
	types := dedupTypes(actionSpec)

	// build resource requests for sensuctl
	for _, t := range types {
		resource, err := Resolve(t)
		if err != nil {
			return nil, fmt.Errorf("invalid resource type: %s", t)
		}
		if lifter, ok := resource.(lifter); ok {
			resource = lifter.Lift()
		}
		actions = append(actions, resource)
	}
	return actions, nil
}

// TrimResources removes all of the resources in the second slice from the
// first slice, if they are in there.
func TrimResources(resources []corev2.Resource, toRemove []corev2.Resource) []corev2.Resource {
	result := make([]corev2.Resource, 0, len(resources))
	for _, resource := range resources {
		var found bool
		for _, remove := range toRemove {
			if reflect.DeepEqual(resource, remove) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, resource)
		}
	}
	return result
}

// WrapResources takes a list of resources and returns a list of wrappers.
func WrapResources(resources []corev2.Resource) []types.Wrapper {
	wrapped := []types.Wrapper{}
	for _, resource := range resources {
		wrapped = append(wrapped, types.WrapResource(resource))
	}
	return wrapped
}
