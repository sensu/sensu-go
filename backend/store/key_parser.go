package store

import "strings"

// SplitKey contains the component parts of a sensu resource key.
type SplitKey struct {
	Root         string
	ResourceType string
	Organization string
	Environment  string
	ResourceName string
}

// ParseResourceKey splits a resource key into its component parts.
// It assumes the following key structure:
//
// /root/resourcetype/organization/environment/resourcename
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
		result.Organization = split[2]
	}
	if len(split) > 3 {
		result.Environment = split[3]
	}
	if len(split) > 4 {
		result.ResourceName = split[4]
	}
	return result
}

func (s SplitKey) String() string {
	b := NewKeyBuilder(s.ResourceType)
	b = b.WithNamespace(Namespace{Org: s.Organization, Env: s.Environment})

	return b.Build(s.ResourceName)
}
