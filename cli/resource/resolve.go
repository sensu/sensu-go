package resource

import (
	"fmt"
	"regexp"
	"strings"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

var (
	// All is all the core resource types and associated sensuctl verbs (non-namespaced resources are intentionally ordered first).
	All = []types.Resource{
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
		&corev2.Hook{},
		&corev2.Mutator{},
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
	Lift() types.Resource
}

var resourceRE = regexp.MustCompile(`(\w+\/v\d+\.)?(\w+)`)

// Resolve resolves a named resource to an empty concrete type.
// The value is boxed within a types.Resource interface value.
func Resolve(resource string) (types.Resource, error) {
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
	return types.ResolveType(apiVersion, typeName)
}

func dedupTypes(arg string) []string {
	types := strings.Split(arg, ",")
	seen := make(map[string]struct{})
	result := make([]string, 0, len(types))
	for _, t := range types {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		result = append(result, t)
	}
	return result
}

// GetResourceRequests gets the resources based on the input.
func GetResourceRequests(actionSpec string) ([]types.Resource, error) {
	// parse the comma separated resource types and match against the defined actions
	if actionSpec == "all" {
		return All, nil
	}
	var actions []types.Resource
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
