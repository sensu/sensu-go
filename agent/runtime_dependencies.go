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
	// Time in seconds we allow for fetching the asset
	fetchTimeout = time.Second * 30
)

type dependencyManager struct {
	agent        *Agent
	dependencies []*runtimeDependency
}

func newDependencyManager(agent *Agent, check *types.Check) *dependencyManager {
	manager := &dependencyManager{
		agent:        agent,
		dependencies: []*runtimeDependency{},
	}

	if check.RuntimeDependencies == nil {
		return manager
	}

	for _, asset := range check.RuntimeDependencies {
		manager.dependencies = append(
			manager.dependencies,
			&runtimeDependency{
				agent: agent,
				asset: &asset,
			},
		)
	}

	return manager
}

func (m *dependencyManager) install() error {
	for _, dep := range m.dependencies {
		if err := dep.install(); err != nil {
			return err
		}
	}

	return nil
}

func (m *dependencyManager) paths() []string {
	paths := []string{}

	for _, dep := range m.dependencies {
		paths = append(paths, dep.path())
	}

	return paths
}

func (m *dependencyManager) injectIntoEnv(env []string) []string {
	for i, e := range env {
		pair := strings.Split(e, "=")
		k, v := pair[0], pair[1]

		injectPaths := func(subDir string) {
			sep := string(filepath.ListSeparator)
			vals := strings.Split(v, sep)
			for _, p := range m.paths() {
				fullpath := filepath.Join(p, subDir)
				vals = append([]string{fullpath}, vals...)
			}
			env[i] = fmt.Sprintf("%s=%s", k, strings.Join(vals, sep))
		}

		switch k {
		case "PATH":
			injectPaths("bin")
		case "LD_LIBRARY_PATH":
			injectPaths("lib")
		case "CPATH":
			injectPaths("include")
		}
	}

	return env
}

type runtimeDependency struct {
	agent *Agent
	asset *types.Asset
}

func (d *runtimeDependency) path() string {
	config := d.agent.config
	return filepath.Join(config.CacheDir, "deps", d.asset.Hash)
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
		return fmt.Errorf("unable to create cache directory '%s'", d.path(), err)
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
