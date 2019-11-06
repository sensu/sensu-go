package cache

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// GetSilencedByEvent returns the name of all Silenced entries that match the given event.
func GetSilencedByEvent(event *corev2.Event, c *Resource) []*corev2.Silenced {
	if !event.HasCheck() {
		return []*corev2.Silenced{}
	}

	resources := c.Get(event.Check.Namespace)
	entries := make([]*corev2.Silenced, len(resources))
	for i, resource := range resources {
		entries[i] = resource.Resource.(*corev2.Silenced)
	}

	return entries
}

// GetSilencedByName returns a Silenced entry matching the given name, or nil if it doesn't exist.
func GetSilencedByName(namespace, name string, c *Resource) *corev2.Silenced {
	resources := c.Get(namespace)
	for _, resource := range resources {
		silenced := resource.Resource.(*corev2.Silenced)
		if silenced.ObjectMeta.Name == name {
			return silenced
		}
	}
	return nil
}
