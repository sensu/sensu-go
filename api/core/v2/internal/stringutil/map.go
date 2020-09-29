package stringutil

// Merge contents of one map into another using a prefix.
func MergeMapWithPrefix(a map[string]string, b map[string]string, prefix string) {
	for k, v := range b {
		a[prefix+k] = v
	}
}
