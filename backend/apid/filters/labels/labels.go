package labels

import (
	corev2 "github.com/sensu/core/v2"
	coreTypes "github.com/sensu/core/v3/types"
	"github.com/sensu/sensu-go/util/compat"
)

// Filter iterates through the given resources and filters, without allocating,
// based on the matches function result. It supports both resources and wrapped
// resources.
func Filter(resources interface{}, matches func(set map[string]string) bool) interface{} {
	switch resources := resources.(type) {
	case []corev2.Resource:
		filteredResources := resources[:0]
		for _, resource := range resources {
			if matches(resource.GetObjectMeta().Labels) {
				filteredResources = append(filteredResources, resource)
			} else if event, ok := resource.(*corev2.Event); ok {
				// If we have an event, we also want to inspect the check and entitiy
				// labels. Events are never wrapped in the store so we do not need this
				// condition in the switch's case below.
				labels := merge(event.Entity.GetLabels(), event.Check.GetLabels())
				if matches(labels) {
					filteredResources = append(filteredResources, resource)
				}
			}
		}
		return filteredResources
	case []*coreTypes.Wrapper:
		filteredResources := resources[:0]
		for _, resource := range resources {
			if matches(compat.GetObjectMeta(resource.Value).Labels) {
				filteredResources = append(filteredResources, resource)
			}
		}
		return filteredResources
	default:
		logger.Debug("unable to filter non-resources and non-wrapped resources")
	}

	return resources
}

func merge(ms ...map[string]string) map[string]string {
	result := map[string]string{}
	for _, m := range ms {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
