package gostatsd

import (
	"sort"
	"strings"
)

// Tags represents a list of tags. Tags can be of two forms:
// 1. "key:value". "value" may contain column(s) as well.
// 2. "tag". No column.
// Each tag's key and/or value may contain characters invalid for a particular backend.
// Backends are expected to handle them appropriately. Different backends may have different sets of valid
// characters so it is undesirable to have restrictions on the input side.
type Tags []string

// StatsdSourceID stores the key used to tag metrics with the origin IP address.
// Should be short to avoid extra hashing and memory overhead for map operations.
const StatsdSourceID = "s"

// String returns a comma-separated string representation of the tags.
func (tags Tags) String() string {
	return strings.Join(tags, ",")
}

// SortedString sorts the tags alphabetically and returns
// a comma-separated string representation of the tags.
// Note that this method may mutate the original object.
func (tags Tags) SortedString() string {
	sort.Strings(tags)
	return tags.String()
}

// NormalizeTagKey cleans up the key of a tag.
func NormalizeTagKey(key string) string {
	return strings.Replace(key, ":", "_", -1)
}

// Concat returns a new Tags with the additional ones added
func (tags Tags) Concat(additional Tags) Tags {
	t := make(Tags, 0, len(tags)+len(additional))
	t = append(t, tags...)
	t = append(t, additional...)
	return t
}

// Copy returns a copy of the Tags
func (tags Tags) Copy() Tags {
	if tags == nil {
		return nil
	}
	tagCopy := make(Tags, len(tags))
	copy(tagCopy, tags)
	return tagCopy
}
