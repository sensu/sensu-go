package api

import "strings"

// IsInternal returns true if the kind name is internal.
func IsInternal(kind string) bool {
	return !strings.Contains(kind, "/")
}
