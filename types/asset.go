package types

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"time"
)

// AssetNameRegex used to validate name of asset
var AssetNameRegex = regexp.MustCompile(`^[a-z0-9\/\_\.\-]+$`)

// Asset defines an asset agents install as a dependency for a check.
type Asset struct {
	// Name is the unique identifier for an asset.
	Name string `json:"name"`

	// Url is the location of the asset.
	URL string `json:"url"`

	// Hash digest of asset
	Hash string `json:"hash"`

	// Metadata is a set of key value pair associated with the asset.
	Metadata map[string]string `json:"metadata"`

	// Organization indicates to which org an asset belongs
	Organization string `json:"organization"`
}

// Validate returns an error if the asset contains invalid values.
func (a *Asset) Validate() error {
	if err := ValidateAssetName(a.Name); err != nil {
		return err
	}

	if a.URL == "" {
		return errors.New("URL cannot be empty")
	}

	if a.Organization == "" {
		return errors.New("organization cannot be empty")
	}

	u, err := url.Parse(a.URL)
	if err != nil {
		return errors.New("invalid URL provided")
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		return errors.New("URL must be HTTP or HTTPS")
	}

	return nil
}

// ValidateAssetName validates that asset's name is valid
func ValidateAssetName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	if !AssetNameRegex.MatchString(name) {
		return errors.New(
			"name must be lowercase and may only contain forward slashes, underscores, dashes and numbers",
		)
	}

	return nil
}

// UpdateHash sets Hash value from given URL
func (a *Asset) UpdateHash() (err error) {
	netClient := &http.Client{Timeout: time.Second * 10}
	h := sha256.New()

	r, err := netClient.Get(a.URL)
	if err != nil {
		return
	}

	if _, err = io.Copy(h, r.Body); err != nil {
		return
	}

	a.Hash = hex.EncodeToString(h.Sum(nil))
	return
}

// Filename returns the filename of the underlying asset; pulled from the URL
func (a *Asset) Filename() string {
	u, err := url.Parse(a.URL)
	if err != nil {
		return ""
	}

	_, file := path.Split(u.EscapedPath())
	return file
}

// FixtureAsset given a name returns a valid asset for use in tests
func FixtureAsset(name string) *Asset {
	bytes := make([]byte, 10)
	rand.Read(bytes)
	hash := hex.EncodeToString(bytes)

	return &Asset{
		Name: name,
		Hash: hash,
		URL:  "https://localhost/" + hash + ".zip",
		Metadata: map[string]string{
			"Content-Type":            "application/zip",
			"X-Intended-Distribution": "trusty-14",
		},
		Organization: "default",
	}
}
