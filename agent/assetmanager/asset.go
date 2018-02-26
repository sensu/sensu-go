package assetmanager

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mholt/archiver"
	"github.com/nightlyone/lockfile"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/eval"
	"github.com/sensu/sensu-go/util/retry"
	filetype "gopkg.in/h2non/filetype.v1"
)

const (
	// time in seconds we allow for fetching the asset
	fetchTimeout = time.Second * 30

	// dependencies cache path
	depsCachePath = "deps"
)

// A RuntimeAsset refers to an asset that is currently in use by the agent.
type RuntimeAsset struct {
	path  string
	asset *types.Asset
}

// NewRuntimeAsset given asset and pathPrefix return new managed asset
func NewRuntimeAsset(asset *types.Asset, pathPrefix string) *RuntimeAsset {
	path := filepath.Join(pathPrefix, asset.Sha512)
	return &RuntimeAsset{path: path, asset: asset}
}

func (d *RuntimeAsset) isRelevantTo(entity types.Entity) (bool, error) {
	params := make(map[string]interface{}, 1)
	params["entity"] = entity

	for _, filter := range d.asset.Filters {
		result, err := eval.Evaluate(filter, params)
		if err != nil {
			return false, err
		}

		if !result {
			return false, nil
		}
	}

	return true, nil
}

func (d *RuntimeAsset) isInstalled() (bool, error) {
	installfile := filepath.Join(d.path, ".installed")

	if info, err := os.Stat(installfile); err != nil {
		return false, nil
	} else if info.IsDir() {
		return true, fmt.Errorf("'%s' is a directory", info.Name())
	}

	return true, nil
}

// Add .installed file to asset directory to help us determine if the asset has
// already been installed.
func (d *RuntimeAsset) markAsInstalled() error {
	installfile := filepath.Join(d.path, ".installed")

	file, err := os.Create(installfile)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte{}) // empty file
	return err
}

// Avoid competing installation of assets
func (d *RuntimeAsset) awaitLock() (*lockfile.Lockfile, error) {
	lockfile, _ := lockfile.New(filepath.Join(d.path, ".lock"))

	// Try to lock the asset directory for purpose of writing
	if err := lockfile.TryLock(); err == nil {
		return &lockfile, nil
	}

	// If we fail to get a lock, retry for 90s
	backoff := retry.ExponentialBackoff{
		InitialDelayInterval: 50 * time.Millisecond,
		MaxDelayInterval:     5 * time.Second,
		MaxElapsedTime:       90 * time.Second,
		Multiplier:           1.5,
	}
	if err := backoff.Retry(func(retry int) (bool, error) {
		if retry != 0 {
			logger.Debugf("attempt to acquire a lock #%d", retry)
		}

		if err := lockfile.TryLock(); err != nil {
			logger.WithError(err).Error("attempt to acquire a lock failed")
			return false, nil
		}

		// At this point, the attempt was successful
		logger.Info("successfully acquired a lock")
		return true, nil
	}); err != nil {
		return nil, err
	}

	return &lockfile, nil
}

func (d *RuntimeAsset) fetch() (*http.Response, error) {
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
// nolint
func (d *RuntimeAsset) install() error {
	// Ensure that cache directory exists before we attempt to write the contents
	// of our asset to it.
	binDir := filepath.Join(d.path, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("unable to create cache directory '%s': %s", d.path, err.Error())
	}

	// Obtain a lock to avoid clobbering competing installs
	lockfile, err := d.awaitLock()
	if err != nil {
		return fmt.Errorf("unable to obtain a lock for asset '%s' in a timely manner", d.asset.Name)
	}
	defer lockfile.Unlock()

	// Check that asset hasn't already been installed
	if cached, err := d.isInstalled(); cached || err != nil {
		return err
	}

	// logger.WithFields(logrus.Fields{
	//	"asset_name": d.asset.Name,
	// }).Info("new dependency encountered; downloading")

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
	h := sha512.New()
	if _, err = io.Copy(h, tmpFile); err != nil {
		return fmt.Errorf("generating checksum for asset failed: %s", err.Error())
	}

	// Check that fetched file's checksum matches given
	responseBodySum := hex.EncodeToString(h.Sum(nil))
	if d.asset.Sha512 != responseBodySum {
		return fmt.Errorf(
			"fetched asset checksum did not match '%s' '%s'",
			d.asset.Sha512,
			responseBodySum,
		)
	}

	// Read header
	header := make([]byte, 262)
	tmpFile.Seek(0, 0)
	if _, err = tmpFile.Read(header); err != nil {
		return fmt.Errorf("unable to read asset header: %s", err)
	}

	// Close tempfile to avoid deadlock
	tmpFile.Close()

	// If file is an archive attempt to extract it
	fileKind, _ := filetype.Match(header)
	switch fileKind.MIME.Value {
	case "application/x-tar":
		if err = archiver.Tar.Open(tmpFile.Name(), d.path); err != nil {
			return fmt.Errorf("unable to extract asset to cache directory w/ err %s", err)
		}
	case "application/gzip":
		if err = archiver.TarGz.Open(tmpFile.Name(), d.path); err != nil {
			return fmt.Errorf("unable to extract asset to cache directory w/ err %s", err)
		}
	default:
		return fmt.Errorf(
			"given file of format '%s' does not appear valid",
			fileKind.MIME.Value,
		)
	}

	// Write .completed file
	d.markAsInstalled()

	// Unlock directory so we allow others others to write again
	lockfile.Unlock()

	return nil
}
