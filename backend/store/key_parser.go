package store

import "strings"

// SplitKey contains the component parts of a sensu resource key.
type SplitKey struct {
	Root         string
	ResourceType string
	Namespace    string
	ResourceName string
}

// ParseResourceKey splits a resource key into its component parts.
// It assumes the following key structure:
//
// /root/resourcetype/namespace/resourcename
// With the leading slash being optional.
func ParseResourceKey(key string) SplitKey {
	var result SplitKey
	split := strings.Split(key, "/")
	if len(split[0]) == 0 {
		split = split[1:]
	}
	if len(split) > 0 {
		result.Root = split[0]
	}
	if len(split) > 1 {
		result.ResourceType = split[1]
	}
	if len(split) > 2 {
		result.Namespace = split[2]
	}
	if len(split) > 3 {
		result.ResourceName = split[3]
	}
	return result
}

func (s SplitKey) String() string {
	b := NewKeyBuilder(s.ResourceType)
	b = b.WithNamespace(s.Namespace)

	return b.Build(s.ResourceName)
}
