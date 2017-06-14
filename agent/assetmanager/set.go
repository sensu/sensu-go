package assetmanager

import (
	"fmt"
	"path/filepath"
	"strings"
)

// A RuntimeAssetSet wraps a set of assets and it's ENV variables
type RuntimeAssetSet struct {
	// env is a copy of the current environment with PATH, LD_LIBRARY_PATH, &
	// CPATH updated to include paths to asset. In this way the check's execution
	// context has access to to reference binary & libraries provided by assets.
	env    []string
	assets []*RuntimeAsset
}

// NewRuntimeAssetSet given set of assets and ENV return new context
func NewRuntimeAssetSet(assets []*RuntimeAsset, baseEnv []string) *RuntimeAssetSet {
	set := &RuntimeAssetSet{assets: assets}
	set.computeEnv(baseEnv)
	return set
}

// InstallAll - ensures that all assets are installed
func (setPtr *RuntimeAssetSet) InstallAll() error {
	for _, asset := range setPtr.assets {
		if err := asset.install(); err != nil {
			return err
		}
	}

	return nil
}

// Env - includes all environment
func (setPtr *RuntimeAssetSet) Env() []string {
	return setPtr.env
}

// Injects PATH, LD_LIBRARY_PATH, & CPATH into given ENV representation
func (setPtr *RuntimeAssetSet) computeEnv(baseEnv []string) {
	assetPaths := setPtr.paths()

	// Instantiate copy of existing environment
	setPtr.env = make([]string, len(baseEnv))
	copy(setPtr.env, baseEnv)

	// Inject paths for assets
	for i, e := range setPtr.env {
		pair := strings.Split(e, "=")
		key, val := pair[0], pair[1]

		switch key {
		case "PATH":
			val = injectPathsIntoEnvVar(assetPaths, val, "bin")
		case "LD_LIBRARY_PATH":
			val = injectPathsIntoEnvVar(assetPaths, val, "lib")
		case "CPATH":
			val = injectPathsIntoEnvVar(assetPaths, val, "include")
		default:
			continue
		}

		setPtr.env[i] = fmt.Sprintf("%s=%s", key, val)
	}
}

func (setPtr *RuntimeAssetSet) paths() []string {
	paths := make([]string, len(setPtr.assets))

	for _, asset := range setPtr.assets {
		paths = append(paths, asset.path)
	}

	return paths
}

func injectPathsIntoEnvVar(paths []string, val, subDir string) string {
	for _, p := range paths {
		val = strings.Join(
			[]string{filepath.Join(p, subDir), val},
			string(filepath.ListSeparator),
		)
	}
	return val
}
