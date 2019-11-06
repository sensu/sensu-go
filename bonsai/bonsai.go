package bonsai

import (
	"errors"
	"sort"
	"strings"

	goversion "github.com/hashicorp/go-version"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// Asset stores information about an asset (metadata, versions, etc.) from Bonsai
type Asset struct {
	// Name is the full name (including namespace) of the asset
	Name string `json:"name,omitempty"`
	// Description is the description for the asset
	Description string `json:"description,omitempty"`
	// URL is the API URL for the asset
	URL string `json:"url,omitempty"`
	// GithubURL is the Github URL for the asset
	GithubURL string `json:"github_url,omitempty"`
	// DownloadURL is the URL to download the asset source code
	DownloadURL string `json:"download_url,omitempty"`
	// Versions is a list of asset versions
	Versions []*AssetVersionGrouping `json:"versions"`
}

// AssetVersionGrouping groups versions to an asset
type AssetVersionGrouping struct {
	// Version is a single version for an asset
	Version string `json:"version,omitempty"`
}

// BaseAsset is the base bonsai asset
type BaseAsset struct {
	// Namespace is the Bonsai asset namespace (username)
	Namespace string `json:"namespace,omitempty"`
	// Name is the Bonsai asset name
	Name string `json:"name,omitempty"`
	// Version is a single version for an asset
	Version string `json:"version,omitempty"`
}

// OutdatedAsset has information about new assets in bonsai
type OutdatedAsset struct {
	// Name is the name of the Bonsai asset
	Name string `json:"bonsai_name,omitempty"`
	// Namespace is the Bonsai namespace (aka username)
	Namespace string `json:"bonsai_namespace,omitempty"`
	// AssetName is the name of the Sensu asset
	AssetName string `json:"asset_name,omitempty"`
	// CurrentVersion is the version of the Sensu asset currently installed
	CurrentVersion string `json:"current_version,omitempty"`
	// LatestVersion is the latest version of the asset in Bonsai
	LatestVersion string `json:"latest_version,omitempty"`
}

// NewBaseAsset creates a new BaseAsset
func NewBaseAsset(name string) (*BaseAsset, error) {
	bAsset := BaseAsset{}

	versionSplit := strings.Split(name, ":")
	if len(versionSplit) > 2 {
		return nil, errors.New("only one colon is permitted in the asset name")
	}
	if len(versionSplit) == 2 {
		bAsset.Version = versionSplit[1]
	}
	nameWithoutVersion := versionSplit[0]

	nameSplit := strings.Split(nameWithoutVersion, "/")
	if len(nameSplit) != 2 {
		return nil, errors.New("asset name must be specified in the format \"namespace/asset\" (e.g. mynamespace/cpu-check)")
	}
	bAsset.Namespace = nameSplit[0]
	bAsset.Name = nameSplit[1]

	return &bAsset, nil
}

// HasVersion will return whether or not a Bonsai asset contains a specific version.
func (b *Asset) HasVersion(version *goversion.Version) bool {
	for _, validVersion := range b.ValidVersions() {
		if validVersion.Equal(version) {
			return true
		}
	}
	return false
}

// ValidVersions will return a list of unsorted semver-compliant versions.
// Any versions that cannot be parsed as semver-compliant will be disregarded.
func (b *Asset) ValidVersions() []*goversion.Version {
	var versions []*goversion.Version

	for _, bVersion := range b.Versions {
		version, err := goversion.NewVersion(bVersion.Version)
		if err == nil {
			versions = append(versions, version)
		}
	}
	return versions
}

// LatestVersion will return the highest semver-compliant version.
func (b *Asset) LatestVersion() *goversion.Version {
	versions := b.ValidVersions()
	sort.Sort(goversion.Collection(versions))
	return versions[len(versions)-1]
}

// GetObjectMeta is just a no-op to satisfy the Resource interface
func (o *OutdatedAsset) GetObjectMeta() corev2.ObjectMeta {
	return corev2.ObjectMeta{}
}

// SetObjectMeta is just a no-op to satisfy the Resource interface
func (o *OutdatedAsset) SetObjectMeta(meta corev2.ObjectMeta) {
	// no-op
}

// SetNamespace is just a no-op to satisfy the Resource interface
func (o *OutdatedAsset) SetNamespace(namespace string) {
	// no-op
}

// StorePrefix is just a no-op to satisfy the Resource interface
func (o *OutdatedAsset) StorePrefix() string {
	return ""
}

// RBACName is just a no-op to satisfy the Resource interface
func (o *OutdatedAsset) RBACName() string {
	return ""
}

// URIPath is just a no-op to satisfy the Resource interface
func (o *OutdatedAsset) URIPath() string {
	return ""
}

// Validate is just a no-op to satisfy the Resource interface
func (o *OutdatedAsset) Validate() error {
	return nil
}
