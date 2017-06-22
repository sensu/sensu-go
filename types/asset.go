package types

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"time"

	"github.com/Knetic/govaluate"
)

// AssetNameRegex used to validate name of asset
var AssetNameRegex = regexp.MustCompile(`^[a-z0-9\/\_\.\-]+$`)

// Asset defines an asset agents install as a dependency for a check.
type Asset struct {
	// Name is the unique identifier for an asset.
	Name string `json:"name"`

	// Url is the location of the asset.
	URL string `json:"url"`

	// Sha512 is the SHA-512 checksum of the asset
	Sha512 string `json:"sha512"`

	// Metadata is a set of key value pair associated with the asset.
	Metadata map[string]string `json:"metadata"`

	// Filters is a collection of sensu queries, used by the system to determine
	// if the asset should be installed. If more than one filter is present the
	// queries are joined by the "AND" operator.
	Filters []string `json:"filters"`

	// Organization indicates to which org an asset belongs
	Organization string `json:"organization"`
}

// Validate returns an error if the asset contains invalid values.
func (a *Asset) Validate() error {
	if err := ValidateAssetName(a.Name); err != nil {
		return err
	}

	if a.Organization == "" {
		return errors.New("organization cannot be empty")
	}

	if a.Sha512 == "" {
		return errors.New("SHA-512 checksum cannot be empty")
	}

	if a.URL == "" {
		return errors.New("URL cannot be empty")
	}

	u, err := url.Parse(a.URL)
	if err != nil {
		return errors.New("invalid URL provided")
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		return errors.New("URL must be HTTP or HTTPS")
	}

	for _, filter := range a.Filters {
		if err := ValidateAssetFilter(filter); err != nil {
			return fmt.Errorf("invalid filter syntax for '%s'", filter)
		}
	}

	return nil
}

// ValidateAssetFilter ensure that the given filter is can be parse successfully
// and that it does not contain any modifier tokens.
func ValidateAssetFilter(filter string) error {
	exp, err := govaluate.NewEvaluableExpression(filter)
	if err != nil {
		return err
	}

	// Do not allow modifier tokens (eg. +, -, /, *, **, &, etc.)
	for _, token := range exp.Tokens() {
		if token.Kind == govaluate.MODIFIER {
			return errors.New("invalid token encountered")
		}
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
		Name:   name,
		Sha512: "25e01b962045f4f5b624c3e47e782bef65c6c82602524dc569a8431b76cc1f57639d267380a7ec49f70876339ae261704fc51ed2fc520513cf94bc45ed7f6e17",
		URL:    "https://localhost/" + hash + ".zip",
		Metadata: map[string]string{
			"Content-Type":            "application/zip",
			"X-Intended-Distribution": "trusty-14",
		},
		Organization: "default",
	}
}
