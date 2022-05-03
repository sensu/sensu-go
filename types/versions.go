package types

import (
	"fmt"
	"path"
	"runtime/debug"
	"strings"

	"github.com/blang/semver/v4"
)

// APIModuleVersions returns a map of Sensu API modules that are compiled into
// the product.
func APIModuleVersions() map[string]string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok || buildInfo.Deps == nil {
		// fallback case - ReadBuildInfo() not available in tests. Remove later.
		return map[string]string{
			"core/v2": "v2.6.0",
			"core/v3": "v3.3.0",
		}
	}
	apiModuleVersions := make(map[string]string)
	packageMapMu.Lock()
	defer packageMapMu.Unlock()
	for k := range packageMap {
		for _, mod := range buildInfo.Deps {
			if strings.HasSuffix(mod.Path, path.Join("api", k)) {
				apiModuleVersions[k] = mod.Version
				break
			}
		}
	}
	return apiModuleVersions
}

// ParseAPIVersion parses an api_version that looks like the following:
//
// core/v2
// core/v2.2
// core/v2.2.1
//
// It returns the name of the apiGroup (core/v2), and the semantic version
// (v2.0.0, v2.2.0, v2.2.1). A leading 'v' is included, keeping with how Go
// modules express their versions.
//
// If ParseAPIVersion can't determine the version, for instance if it's passed
// a string that does not seem to be a versioned API group, it will return its
// input as the apiGroup, and v0.0.0 as the version.
func ParseAPIVersion(apiVersion string) (apiGroup, semVer string) {
	group, version := path.Split(apiVersion)
	if version == "" {
		// There is no version for the API group, which is fine.
		return group, "v0.0.0"
	}
	semver, err := semver.ParseTolerant(version)
	if err != nil {
		// It's not the expected format
		return apiVersion, "v0.0.0"
	}
	apiGroup = path.Join(group, fmt.Sprintf("v%d", semver.Major))
	semVer = fmt.Sprintf("v%d.%d.%d", semver.Major, semver.Minor, semver.Patch)
	return apiGroup, semVer
}
