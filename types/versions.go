package types

import (
	"path"
	"runtime/debug"
	"strings"
)

// APIModuleVersions returns a map of Sensu API modules that are compiled into
// the product.
func APIModuleVersions() map[string]string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
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
