package v2

import "path"

const (
	// APIGroupName is the group name for this API
	APIGroupName = "core"

	// APIVersion is the version for this API
	APIVersion = "v2"
)

// URLPrefix is the URL prefix of this API
var URLPrefix = path.Join("/api", APIGroupName, APIVersion)
