package v2

import (
	"errors"
	"sort"
	"strings"

	goversion "github.com/hashicorp/go-version"
)

// NewBonsaiBaseAsset creates a new BonsaiBaseAsset
func NewBonsaiBaseAsset(name string) (*BonsaiBaseAsset, error) {
	bAsset := BonsaiBaseAsset{}

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

// ContainsVersion will return whether or not a Bonsai asset contains a specific version.
func (b *BonsaiAsset) HasVersion(version *goversion.Version) bool {
	for _, validVersion := range b.ValidVersions() {
		if validVersion.Equal(version) {
			return true
		}
	}
	return false
}

// GetValidVersions will return a list of unsorted semver-compliant versions.
// Any versions that cannot be parsed as semver-compliant will be disregarded.
func (b *BonsaiAsset) ValidVersions() []*goversion.Version {
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
func (b *BonsaiAsset) LatestVersion() *goversion.Version {
	versions := b.ValidVersions()
	sort.Sort(goversion.Collection(versions))
	return versions[len(versions)-1]
}

// GetObjectMeta is just a no-op to satisfy the Resource interface
func (o *OutdatedBonsaiAsset) GetObjectMeta() ObjectMeta {
	return ObjectMeta{}
}

// SetObjectMeta is just a no-op to satisfy the Resource interface
func (o *OutdatedBonsaiAsset) SetObjectMeta(meta ObjectMeta) {
	// no-op
}

// SetNamespace is just a no-op to satisfy the Resource interface
func (o *OutdatedBonsaiAsset) SetNamespace(namespace string) {
	// no-op
}

// StorePrefix is just a no-op to satisfy the Resource interface
func (o *OutdatedBonsaiAsset) StorePrefix() string {
	return ""
}

// RBACName is just a no-op to satisfy the Resource interface
func (o *OutdatedBonsaiAsset) RBACName() string {
	return ""
}

// URIPath is just a no-op to satisfy the Resource interface
func (o *OutdatedBonsaiAsset) URIPath() string {
	return ""
}

// Validate is just a no-op to satisfy the Resource interface
func (o *OutdatedBonsaiAsset) Validate() error {
	return nil
}
