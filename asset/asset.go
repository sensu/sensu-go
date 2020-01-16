// Package asset provides a mechanism for installing, managing, and utilizing
// Sensu Assets.
//
// Access to assets are serialized. When an asset is first encountered,
// getting the asset from the manager blocks until the asset has been
// fetched, verified, and expanded on the host filesystem (or deemed
// unnecessary due to asset filters).
//
// The first goroutine to get an asset will cause the installation, and
// subsequent calling goroutines will simply block while installation
// completes. If the initial installation fails, the next goroutine to
// unblock will attempt reinstallation.
package asset

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const (
	binDir     = "bin"
	libDir     = "lib"
	includeDir = "include"
)

// A Getter is responsible for fetching (based on filter selection), verifying,
// and expanding an asset. Calls to the Get method block until the Asset has
// fetched, verified, and expanded or it returns an error indicating why getting
// the asset failed.
//
// If the context is canceled while Get is in progress, then the operation will
// be canceled and the error from the context will be returned.
type Getter interface {
	Get(context.Context, *corev2.Asset) (*RuntimeAsset, error)
}

// A RuntimeAsset is a locally expanded Asset.
type RuntimeAsset struct {
	// Path is the absolute path to the asset's base directory.
	Path string
	// SHA512 is the hash of the asset tarball.
	SHA512 string
}

// BinDir returns the full path to the asset's bin directory.
func (r *RuntimeAsset) BinDir() string {
	return filepath.Join(r.Path, binDir)
}

// LibDir returns the full path to the asset's lib directory.
func (r *RuntimeAsset) LibDir() string {
	return filepath.Join(r.Path, libDir)
}

// IncludeDir returns the full path to the asset's include directory.
func (r *RuntimeAsset) IncludeDir() string {
	return filepath.Join(r.Path, includeDir)
}

// Env returns a list of environment variables (e.g. 'PATH=...', 'CPATH=...')
// with asset-specific paths prepended to the parent environment paths for
// each variable, allowing an asset to be used during execution.
func (r *RuntimeAsset) Env() []string {
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
func (r *RuntimeAsset) joinPaths(pathFunc func(*RuntimeAsset) string) string {
	var sb strings.Builder
	sb.WriteString(pathFunc(r))
	sb.WriteRune(os.PathListSeparator)
	return sb.String()
}
