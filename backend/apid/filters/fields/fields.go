package fields

import (
	corev3 "github.com/sensu/core/v3"
	coreTypes "github.com/sensu/sensu-go/types"
)

type FieldsFunc func(resource corev3.Resource) map[string]string

// Filter iterates through the given resources and filters, without allocating,
// based on the matches function result. It supports both resources and wrapped
// resources.
func Filter(resources interface{}, matches func(set map[string]string) bool, fields FieldsFunc) interface{} {
	switch resources := resources.(type) {
	case []corev3.Resource:
		filteredResources := resources[:0]
		for _, resource := range resources {
			if matches(fields(resource)) {
				filteredResources = append(filteredResources, resource)
			}
		}
		return filteredResources
	case []*coreTypes.Wrapper:
		filteredResources := resources[:0]
		for _, resource := range resources {
			if matches(fields(resource.Value.(corev3.Resource))) {
				filteredResources = append(filteredResources, resource)
			}
		}
		return filteredResources
	}

	return resources
}
