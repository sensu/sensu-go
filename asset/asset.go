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
	"path/filepath"

	"github.com/sensu/sensu-go/types"
)

const (
	binDir     = "bin"
	libDir     = "lib"
	includeDir = "include"
)

// A Getter is responsible for fetching (based on fitler selection), verifying,
// and expanding an asset. Calls to the Get method block until the Asset has
// fetched, verified, and expanded or it returns an error indicating why getting
// the asset failed.
type Getter interface {
	Get(*types.Asset) (*RuntimeAsset, error)
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
