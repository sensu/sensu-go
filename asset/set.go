package asset

import (
	"fmt"
	"os"
	"strings"

	"github.com/sensu/sensu-go/types"
)

// RuntimeAssetSet is a set of runtime assets.
type RuntimeAssetSet []*RuntimeAsset

// GetAll gets a list of assets with the provided getter.
func GetAll(getter Getter, assets []types.Asset) (RuntimeAssetSet, error) {
	runtimeAssets := make([]*RuntimeAsset, 0, len(assets))
	for _, asset := range assets {
		runtimeAsset, err := getter.Get(&asset)
		if err != nil {
			return nil, err
		}
		if runtimeAsset != nil {
			runtimeAssets = append(runtimeAssets, runtimeAsset)
		}
	}
	return runtimeAssets, nil
}

// Env returns a list of environment variables (e.g. 'PATH=...', 'CPATH=...')
// with asset-specific paths prepended to the parent environment paths for
// each variable, allowing an asset to be used during check execution.
func (r *RuntimeAssetSet) Env() []string {
	assetEnv := []string{
		fmt.Sprintf("PATH=%s${PATH}", r.joinPaths((*RuntimeAsset).BinDir)),
		fmt.Sprintf("LD_LIBRARY_PATH=%s${LD_LIBRARY_PATH}", r.joinPaths((*RuntimeAsset).LibDir)),
		fmt.Sprintf("CPATH=%s${CPATH}", r.joinPaths((*RuntimeAsset).IncludeDir)),
	}
	for i, envVar := range assetEnv {
		// ExpandEnv replaces ${var} with the contents of var from the current
		// environment, or an empty string if var doesn't exist.
		assetEnv[i] = os.ExpandEnv(envVar)
	}
	return assetEnv
}

// joinPaths joins all paths of a given type for each asset in RuntimeAssetSet.
func (r *RuntimeAssetSet) joinPaths(pathFunc func(*RuntimeAsset) string) string {
	var sb strings.Builder
	for _, asset := range *r {
		sb.WriteString(pathFunc(asset))
		sb.WriteRune(os.PathListSeparator)
	}
	return sb.String()
}
