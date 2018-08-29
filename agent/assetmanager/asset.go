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
	"github.com/sirupsen/logrus"
	filetype "gopkg.in/h2non/filetype.v1"
	filetype_types "gopkg.in/h2non/filetype.v1/types"
)

const (
	// time in seconds we allow for fetching the asset
	fetchTimeout = time.Second * 30

	// dependencies cache path
	depsCachePath = "deps"

	// Size of file header for sniffing type
	headerSize = 262
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
		result, err := eval.EvaluatePredicate(filter, params)
		if err != nil {
			switch err.(type) {
			case eval.SyntaxError, eval.TypeError:
				return false, err
			default:
				// Other errors during execution are likely due to missing attrs,
				// simply continue in this case.
				continue
			}
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

	// empty file
	return file.Close()
}

// Avoid competing installation of assets
func (d *RuntimeAsset) awaitLock() (*lockfile.Lockfile, error) {
	lockpath, err := filepath.Abs(filepath.Join(d.path, ".lock"))
	if err != nil {
		return nil, err
	}

	lockfile, err := lockfile.New(lockpath)
	if err != nil {
		return nil, err
	}

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

// binDir creates the asset's bin directory and returns the path
func (d *RuntimeAsset) binDir() (string, error) {
	// Ensure that cache directory exists before we attempt to write the contents
	// of our asset to it.
	binDir := filepath.Join(d.path, "bin")
	err := os.MkdirAll(binDir, os.ModeDir|0700)
	if err != nil {
		err = fmt.Errorf("error creating directory %q: %s", binDir, err)
	}
	return binDir, err
}

func (d *RuntimeAsset) download() (*os.File, error) {
	// Download the asset
	r, err := d.fetch()
	if err != nil {
		return nil, err
	}

	// Write response to tmp
	tmpFile, err := ioutil.TempFile(os.TempDir(), "sensu-asset")
	if err != nil {
		return nil, fmt.Errorf("can't open tmp file for asset %q", d.asset.Name)
	}

	if _, err = io.Copy(tmpFile, r.Body); err != nil {
		return nil, fmt.Errorf("error downloading asset %q", d.asset.Name)
	}

	return tmpFile, resetFile(tmpFile)
}

func hashFile(f *os.File) (string, error) {
	// Generate checksum for downloaded file
	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("generating checksum for asset failed: %s", err)
	}

	return hex.EncodeToString(h.Sum(nil)), resetFile(f)
}

func sniffType(f *os.File) (filetype_types.Type, error) {
	header := make([]byte, headerSize)
	if _, err := f.Read(header); err != nil {
		return filetype_types.Type{}, fmt.Errorf("unable to read asset header: %s", err)
	}
	ft, err := filetype.Match(header)
	if err != nil {
		return ft, err
	}
	return ft, resetFile(f)
}

func resetFile(f *os.File) error {
	// Ensure file contents are synced and rewound
	if err := f.Sync(); err != nil {
		return err
	}
	_, err := f.Seek(0, 0)
	return err
}

// Downloads the given depdencies asset to the cache directory.
func (d *RuntimeAsset) install() (err error) {
	if _, err := d.binDir(); err != nil {
		return err
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

	logger.WithFields(logrus.Fields{
		"asset": d.asset.Name,
	}).Info("downloading asset")

	// Download the asset
	tmpFile, err := d.download()
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Generate checksum for downloaded file
	checksum, err := hashFile(tmpFile)
	if err != nil {
		return err
	}

	// validate checksum
	if d.asset.Sha512 != checksum {
		return fmt.Errorf("asset checksum does not match: %q != %q", d.asset.Sha512, checksum)
	}

	// detect the type of archive the asset is
	ft, err := sniffType(tmpFile)
	if err != nil {
		return err
	}

	var ar archiver.Archiver

	// If the file is not an archive, exit with an error.
	switch ft.MIME.Value {
	case "application/x-tar":
		ar = archiver.Tar
	case "application/gzip":
		ar = archiver.TarGz
	default:
		return fmt.Errorf(
			"given file of format '%s' does not appear valid",
			ft.MIME.Value,
		)
	}

	// Extract the archive to the desired path
	if err := ar.Read(tmpFile, d.path); err != nil {
		return fmt.Errorf("error extracting asset: %s", err)
	}

	// Write .completed file
	if err := d.markAsInstalled(); err != nil {
		return fmt.Errorf("error finalizing asset installation: %s", err)
	}

	return nil
}
