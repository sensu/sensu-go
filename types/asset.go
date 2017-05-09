package types

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
)

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
}

// Validate returns an error if the asset contains invalid values.
func (a *Asset) Validate() error {
	if a.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if a.URL == "" {
		return errors.New("URL cannot be empty")
	}

	if a.Hash == "" {
		return errors.New("Hash cannot be empty")
	}

	u, err := url.Parse(a.URL)
	if err != nil {
		return errors.New("Invalid URL provided")
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		return errors.New("URL must be HTTP or HTTPS")
	}

	return nil
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
	}
}
