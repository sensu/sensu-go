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
)

const (
	binDir     = "bin"
	libDir     = "lib"
	includeDir = "include"
)

// An RuntimeAsset is a locally expanded Asset. After downloading, verifying,
// and expanding the Asset, the RuntimeAsset struct contains everything necessary
// to create a runtime environment composed of one or more RuntimeAssets.
type RuntimeAsset struct {

	// The fully-qualified local path to the asset.
	Path string
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
