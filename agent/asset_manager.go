package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/mholt/archiver"
	"github.com/sensu/sensu-go/types"
)

const (
	// time in seconds we allow for fetching the asset
	fetchTimeout = time.Second * 30

	// dependencies cache path
	depsCachePath = "deps"
)

// AssetManager manages caching & installation of dependencies
type AssetManager struct {
	// BaseEnv env vars that dependencies are injected into; defaults to sys vars
	BaseEnv []string

	cacheDir  string
	knownDeps map[string]*runtimeDependency
	env       []string
}

// NewAssetManager given agent returns instantiated AssetManager
func NewAssetManager(agentCacheDir string) *AssetManager {
	manager := &AssetManager{}
	manager.SetCacheDir(agentCacheDir)
	manager.Reset()

	return manager
}

// SetCacheDir sets cache directory given a base directory
func (m *AssetManager) SetCacheDir(baseDir string) {
	m.cacheDir = filepath.Join(baseDir, depsCachePath)
	m.env = nil // Clear environment as it is likely pointing to the wrong paths
}

// Merge updates list of known dependencies given a list of assets
func (m *AssetManager) Merge(assets []types.Asset) {
	for _, asset := range assets {
		// Do nothing if we already know about the dependency
		if m.knownDeps[asset.Hash] != nil {
			continue
		}

		m.env = nil // Clears env forcing us to re-construct on next exec
		m.knownDeps[asset.Hash] = &runtimeDependency{
			manager: m,
			asset:   &asset,
		}
	}
}

// Install - ensures that all known dependencies are installed
func (m *AssetManager) Install() error {
	for _, dep := range m.knownDeps {
		if err := dep.install(); err != nil {
			return err
		}
	}

	return nil
}

// Env returns a copy of the current environment with PATH, LD_LIBRARY_PATH, &
// CPATH updated to include dependency paths. In this way the execution context
// allows the check to reference binary & libraries provided by dependencies.
func (m *AssetManager) Env() []string {
	if m.env != nil {
		return m.env
	}

	injectPaths := func(val string, subDir string) string {
		for _, p := range m.paths() {
			val = strings.Join(
				[]string{filepath.Join(p, subDir), val},
				string(filepath.ListSeparator),
			)
		}
		return val
	}

	// NOTE: Because we memoize this means that if a new ENVIRONMENT variable
	// variable is added it will not be made available in the execution context.
	// In the future we may want have functionality to reset state; perhaps the
	// agent responds to kill -HUP?
	m.env = make([]string, len(m.BaseEnv))
	copy(m.env, m.BaseEnv)

	// Inject paths for dependencies
	for i, e := range m.env {
		pair := strings.Split(e, "=")
		key, val := pair[0], pair[1]

		switch key {
		case "PATH":
			val = injectPaths(val, "bin")
		case "LD_LIBRARY_PATH":
			val = injectPaths(val, "lib")
		case "CPATH":
			val = injectPaths(val, "include")
		default:
			continue
		}

		m.env[i] = fmt.Sprintf("%s=%s", key, val)
	}

	return m.Env()
}

// Reset clears all knownDeps and env from state, this forces the agent to
// recompute the next time a check is run.
//
// NOTE: Cache on disk is not cleared.
func (m *AssetManager) Reset() {
	m.knownDeps = map[string]*runtimeDependency{}
	m.BaseEnv = os.Environ()
	m.env = nil
}

func (m *AssetManager) paths() []string {
	paths := []string{}

	for _, dep := range m.knownDeps {
		paths = append(paths, dep.path())
	}

	return paths
}

type runtimeDependency struct {
	manager *AssetManager
	asset   *types.Asset
}

func (d *runtimeDependency) path() string {
	return filepath.Join(
		d.manager.cacheDir,
		d.asset.Hash,
	)
}

func (d *runtimeDependency) isCached() (bool, error) {
	if info, err := os.Stat(d.path()); err != nil {
		return false, nil
	} else if !info.IsDir() {
		return true, fmt.Errorf("'%s' is not a directory", info.Name())
	}

	return true, nil
}

func (d *runtimeDependency) fetch() (*http.Response, error) {
	// GET asset w/ timeout
	netClient := &http.Client{Timeout: fetchTimeout}
	r, err := netClient.Get(d.asset.URL)
	if err != nil {
		return r, fmt.Errorf("error fetching asset: %s", err.Error())
	}

	return r, err
}

// Downloads the given depdencies asset to the cache directory.
// TODO(james): ugly; too many responsibilities
func (d *runtimeDependency) install() error {
	// Check that asset hasn't already been retrieved
	if cached, err := d.isCached(); cached || err != nil {
		return err
	}

	logger.WithFields(logrus.Fields{
		"asset_name": d.asset.Name,
	}).Info("new dependency encountered; downloading")

	// Download the asset
	r, err := d.fetch()
	if err != nil {
		return err
	}

	// Write response to tmp
	tmpFile, err := ioutil.TempFile(os.TempDir(), "sensu-asset")
	if err != nil {
		return fmt.Errorf("unable to obtain tmp file for asset '%s'", d.asset.Name)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = io.Copy(tmpFile, r.Body); err != nil {
		return fmt.Errorf("unable to write asset '%s' to tmp", d.asset.Name)
	}

	// Ensure file contents are synced and rewound
	tmpFile.Sync()
	tmpFile.Seek(0, 0)

	// Generate checksum for downloaded file
	h := sha256.New()
	if _, err = io.Copy(h, tmpFile); err != nil {
		return fmt.Errorf("generating checksum for asset failed: %s", err.Error())
	}

	// Check that fetched file's checksum matches given
	responseBodySum := hex.EncodeToString(h.Sum(nil))
	if d.asset.Hash != responseBodySum {
		return fmt.Errorf(
			"fetched asset did not match '%s' '%s'",
			d.asset.Hash,
			responseBodySum,
		)
	}

	// Ensure file is synced and closed before we try to extract or move it.
	tmpFile.Close()

	// Ensure that cache directory exists before we attempt to write the contents
	// of our asset to it.
	binDir := filepath.Join(d.path(), "bin")
	if err = os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("unable to create cache directory '%s': %s", d.path(), err.Error())
	}

	// If file is an archive attempt to extract it
	// NOTE(james): For demo purposes, super naive. Having this feature probably
	// doesn't event make sense for the prod release..
	switch r.Header.Get("Content-Type") {
	case "application/x-tar":
		if err = archiver.Tar.Open(tmpFile.Name(), d.path()); err != nil {
			return fmt.Errorf("Unable to extract asset to cache directory. %s", err)
		}
	default:
		filename := filepath.Join(binDir, d.asset.Filename())
		if err = os.Rename(tmpFile.Name(), filename); err != nil {
			return fmt.Errorf("Unable to copy asset to cache directory. %s", err)
		}
	}

	return nil
}
