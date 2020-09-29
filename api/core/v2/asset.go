package v2

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/sensu/sensu-go/api/core/v2/internal/js"
	stringsutil "github.com/sensu/sensu-go/api/core/v2/internal/stringutil"
)

const (
	// AssetsResource is the name of this resource type
	AssetsResource = "assets"
)

var (
	// AssetNameRegexStr used to validate name of asset
	AssetNameRegexStr = `[\w\/\_\.\-\:]+`

	// AssetNameRegex used to validate name of asset
	AssetNameRegex = regexp.MustCompile("^" + AssetNameRegexStr + "$")
)

// StorePrefix returns the path prefix to this resource in the store
func (a *Asset) StorePrefix() string {
	return AssetsResource
}

// URIPath returns the path component of an asset URI.
func (a *Asset) URIPath() string {
	if a.Namespace == "" {
		return path.Join(URLPrefix, AssetsResource, url.PathEscape(a.Name))
	}
	return path.Join(URLPrefix, "namespaces", url.PathEscape(a.Namespace), AssetsResource, url.PathEscape(a.Name))
}

// Validate returns an error if the asset contains invalid values.
func (a *Asset) Validate() error {
	if err := ValidateAssetName(a.Name); err != nil {
		return err
	}

	if a.Namespace == "" {
		return errors.New("namespace cannot be empty")
	}

	if len(a.Builds) == 0 {
		if a.Sha512 == "" {
			return errors.New("SHA-512 checksum cannot be empty")
		}

		if len(a.Sha512) < 128 {
			return errors.New("SHA-512 checksum must be at least 128 characters")
		}

		if a.URL == "" {
			return errors.New("URL cannot be empty")
		}

		return js.ParseExpressions(a.Filters)
	}
	for _, build := range a.Builds {
		if err := build.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate returns an error if the asset contains invalid values.
func (a *AssetBuild) Validate() error {
	if a.Sha512 == "" {
		return errors.New("SHA-512 checksum cannot be empty")
	}

	if len(a.Sha512) < 128 {
		return errors.New("SHA-512 checksum must be at least 128 characters")
	}

	if a.URL == "" {
		return errors.New("URL cannot be empty")
	}

	return js.ParseExpressions(a.Filters)
}

// ValidateAssetName validates that asset's name is valid
func ValidateAssetName(name string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}

	if !AssetNameRegex.MatchString(name) {
		return errors.New(
			"name may only contain letters, forward slashes, underscores, dashes and numbers",
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
	_, _ = rand.Read(bytes)
	hash := hex.EncodeToString(bytes)

	asset := &Asset{
		ObjectMeta: NewObjectMeta(name, "default"),
		Sha512:     "25e01b962045f4f5b624c3e47e782bef65c6c82602524dc569a8431b76cc1f57639d267380a7ec49f70876339ae261704fc51ed2fc520513cf94bc45ed7f6e17",
		URL:        "https://localhost/" + hash + ".zip",
	}
	return asset
}

// NewAsset creates a new Asset.
func NewAsset(meta ObjectMeta) *Asset {
	return &Asset{ObjectMeta: meta}
}

// AssetFields returns a set of fields that represent that resource
func AssetFields(r Resource) map[string]string {
	resource := r.(*Asset)
	fields := map[string]string{
		"asset.name":      resource.ObjectMeta.Name,
		"asset.namespace": resource.ObjectMeta.Namespace,
		"asset.filters":   strings.Join(resource.Filters, ","),
	}
	stringsutil.MergeMapWithPrefix(fields, resource.ObjectMeta.Labels, "asset.labels.")
	return fields
}

// SetNamespace sets the namespace of the resource.
func (a *Asset) SetNamespace(namespace string) {
	a.Namespace = namespace
}

func (a *Asset) RBACName() string {
	return "assets"
}

// SetObjectMeta sets the meta of the resource.
func (a *Asset) SetObjectMeta(meta ObjectMeta) {
	a.ObjectMeta = meta
}
