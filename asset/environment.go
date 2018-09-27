package asset

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	binDir     = "bin"
	libDir     = "lib"
	includeDir = "include"
)

// Env returns a list of environment variables (e.g. 'PATH=...', 'CPATH=...')
// with asset-specific paths prepended to the parent environment paths for
// each variable, allowing an asset to be used during check execution.
func (r *RuntimeAsset) Env() []string {
	sep := string(os.PathListSeparator)
	assetEnv := []string{
		fmt.Sprintf("PATH=%s%s${PATH}", r.BinDir(), sep),
		fmt.Sprintf("LD_LIBRARY_PATH=%s%s${LD_LIBRARY_PATH}", r.LibDir(), sep),
		fmt.Sprintf("CPATH=%s%s", r.IncludeDir(), sep),
	}
	for i, envVar := range assetEnv {
		assetEnv[i] = os.ExpandEnv(envVar)
	}
	return assetEnv
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
