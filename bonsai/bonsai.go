package bonsai

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	goversion "github.com/hashicorp/go-version"
	corev2 "github.com/sensu/core/v2"
)

const (
	// NameAnnotation represents a Bonsai asset name
	NameAnnotation = "io.sensu.bonsai.name"

	// NamespaceAnnotation represents a Bonsai asset namenamespace
	NamespaceAnnotation = "io.sensu.bonsai.namespace"

	// URLAnnotation represents the Bonsai API URL
	URLAnnotation = "io.sensu.bonsai.api_url"

	// VersionAnnotation represents a Bonsai asset version
	VersionAnnotation = "io.sensu.bonsai.version"
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
	// BonsaiName is the name of the Bonsai asset
	BonsaiName string `json:"bonsai_name,omitempty"`
	// BonsaiNamespace is the Bonsai namespace (aka username)
	BonsaiNamespace string `json:"bonsai_namespace,omitempty"`
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

// BonsaiVersion returns the Bonsai version for a given requested version, and
// fallbacks to the latest version if the requested version is nil. An error is
// returned if the requested version does not exist
func (b *Asset) BonsaiVersion(version *goversion.Version) (*goversion.Version, error) {
	if version == nil {
		// No version was requested, therefore return the latest version
		v := b.LatestVersion()
		if v == nil {
			return nil, fmt.Errorf("asset %q does not have any available versions", b.Name)
		}

		return v, nil
	} else if ok, v := b.HasVersion(version); ok {
		// The request version exists, but return the *goversion.Version provided by
		// Bonsai, since its original format might differ from the requested version
		// (e.g. v0.4.0 != 0.4.0), so we don't get a 404
		return v, nil
	}

	// The version requested does not exists. List the available versions and
	// return an error
	availableVersions := b.ValidVersions()
	sort.Sort(goversion.Collection(availableVersions))
	availableVersionStrs := []string{}
	for _, v := range availableVersions {
		availableVersionStrs = append(availableVersionStrs, v.Original())
	}
	return nil, fmt.Errorf("version %q of asset %q does not exist\navailable versions: %s",
		version, b.Name, strings.Join(availableVersionStrs, ", "))
}

// HasVersion will return whether or not a Bonsai asset contains a specific version.
func (b *Asset) HasVersion(version *goversion.Version) (bool, *goversion.Version) {
	for _, validVersion := range b.ValidVersions() {
		if validVersion.Equal(version) {
			return true, validVersion
		}
	}
	return false, nil
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
	if len(versions) == 0 {
		return nil
	}

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
