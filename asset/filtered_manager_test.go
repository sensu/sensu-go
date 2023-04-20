package asset

import (
	"context"
	"errors"
	"fmt"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
)

type MockGetter struct {
	getCalled bool
	asset     *RuntimeAsset
	err       error
}

// Get satisfies the asset.Getter interface
func (m *MockGetter) Get(ctx context.Context, asset *corev2.Asset) (*RuntimeAsset, error) {
	m.getCalled = true
	m.asset.SHA512 = asset.Sha512
	return m.asset, m.err
}

// NewTestFilteredManager creates a new FilteredManager for testing
func NewTestFilteredManager() (*MockGetter, *corev2.Entity, *filteredManager) {
	mockGetter := &MockGetter{asset: &RuntimeAsset{Path: "/foo/bar"}}
	entity := corev2.FixtureEntity("test-entity")
	filteredManager := NewFilteredManager(mockGetter, entity)
	return mockGetter, entity, filteredManager
}

// FilteredManager should call the underlying Getter.
func TestFilteredManagerCallsGetter(t *testing.T) {
	mockGetter, _, filteredManager := NewTestFilteredManager()

	actualAsset, err := filteredManager.Get(context.TODO(), corev2.FixtureAsset("test-asset"))
	assert.NoError(t, err)
	assert.Equal(t, mockGetter.asset, actualAsset)
	assert.True(t, mockGetter.getCalled)
}

// FilteredManager should call underlying Getter on filtered asset.
func TestFilteredManagerFilteredAsset(t *testing.T) {
	mockGetter, entity, filteredManager := NewTestFilteredManager()

	fixtureAsset := corev2.FixtureAsset("test-asset")
	fixtureAsset.Filters = []string{fmt.Sprintf("entity.name == '%s'", entity.Name)}
	actualAsset, err := filteredManager.Get(context.TODO(), fixtureAsset)
	assert.NoError(t, err)
	assert.Equal(t, mockGetter.asset, actualAsset)
	assert.True(t, mockGetter.getCalled)
}

// FilteredManager should not call underlying Getter on unfiltered asset.
func TestFilteredManagerUnfilteredAsset(t *testing.T) {
	mockGetter, _, filteredManager := NewTestFilteredManager()

	fixtureAsset := corev2.FixtureAsset("test-asset")
	fixtureAsset.Filters = []string{"entity.name == 'foo'"}
	actualAsset, err := filteredManager.Get(context.TODO(), fixtureAsset)
	assert.NoError(t, err)
	assert.Nil(t, actualAsset)
	assert.False(t, mockGetter.getCalled)
}

// FilteredManager should pass a build asset if the asset has builds.
// This test ensures that filteredManager detects that an asset build
// exists and passes the build asset to filtereManager's getter
// instead of the asset containing the asset build.
func TestFilteredManagerFilteredBuildAsset(t *testing.T) {
	mockGetter, entity, filteredManager := NewTestFilteredManager()

	url := "http://asset-build-url"
	sha512 := "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
	filters := []string{fmt.Sprintf("entity.name == '%s'", entity.Name)}

	fixtureAsset := corev2.FixtureAsset("test-asset")
	fixtureAsset.Filters = []string{"entity.name == 'foo'"}
	fixtureAsset.Builds = []*corev2.AssetBuild{
		{
			URL:     url,
			Sha512:  sha512,
			Filters: filters,
		},
	}

	actualAsset, err := filteredManager.Get(context.TODO(), fixtureAsset)
	assert.NoError(t, err)
	assert.Equal(t, mockGetter.asset, actualAsset)
	assert.Equal(t, mockGetter.asset.SHA512, sha512)
	assert.NotEqual(t, mockGetter.asset.SHA512, fixtureAsset.Sha512)
	assert.True(t, mockGetter.getCalled)
}

// TestFilteredManagerUnfilteredBuildAsset tests to ensure no asset is returned
// when all build filters do not pass.
func TestFilteredManagerUnfilteredBuildAsset(t *testing.T) {
	_, _, filteredManager := NewTestFilteredManager()

	url := "http://asset-build-url"
	sha512 := "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
	filters := []string{"entity.name == 'asdf'"}

	fixtureAsset := corev2.FixtureAsset("test-asset")
	fixtureAsset.Filters = []string{"entity.name == 'foo'"}
	fixtureAsset.Builds = []*corev2.AssetBuild{
		{
			URL:     url,
			Sha512:  sha512,
			Filters: filters,
		},
	}

	actualAsset, err := filteredManager.Get(context.TODO(), fixtureAsset)
	assert.NoError(t, err)
	assert.Nil(t, actualAsset)
}

// FilteredManager should return error passed by underlying Getter.
func TestFilteredManagerError(t *testing.T) {
	mockGetter, _, filteredManager := NewTestFilteredManager()
	mockGetter.err = errors.New("TestFilteredManagerError")

	_, err := filteredManager.Get(context.TODO(), corev2.FixtureAsset("test-asset"))
	assert.Error(t, err)
	assert.True(t, mockGetter.getCalled)
}

// isFiltered should allow filtering by entity
func TestIsFiltered(t *testing.T) {
	_, entity, filteredManager := NewTestFilteredManager()

	// filtered is true, filter matches
	fixtureAsset := corev2.FixtureAsset("test-asset")
	fixtureAsset.Filters = []string{fmt.Sprintf("entity.name == '%s'", entity.Name)}
	filtered, err := filteredManager.isFiltered(fixtureAsset)
	assert.NoError(t, err)
	assert.True(t, filtered)

	// filtered is true, filter matches
	fixtureAsset.Filters = []string{fmt.Sprintf("entity.system.arch == '%s'", entity.System.Arch)}
	filtered, err = filteredManager.isFiltered(fixtureAsset)
	assert.NoError(t, err)
	assert.True(t, filtered)

	// filtered is false, filter does not match
	fixtureAsset.Filters = []string{fmt.Sprintf("entity.system.arch == '%s'", "foo")}
	filtered, err = filteredManager.isFiltered(fixtureAsset)
	assert.NoError(t, err)
	assert.False(t, filtered)

	// filtered is true, all filters match
	fixtureAsset.Filters = []string{fmt.Sprintf("entity.name == '%s'", entity.Name), "entity.entity_class == 'host'"}
	filtered, err = filteredManager.isFiltered(fixtureAsset)
	assert.NoError(t, err)
	assert.True(t, filtered)

	// filtered is true, filters empty
	fixtureAsset.Filters = []string{}
	filtered, err = filteredManager.isFiltered(fixtureAsset)
	assert.NoError(t, err)
	assert.True(t, filtered)

	// filtered is false, filter does not match
	fixtureAsset.Filters = []string{"entity.name == 'foo'"}
	filtered, err = filteredManager.isFiltered(fixtureAsset)
	assert.NoError(t, err)
	assert.False(t, filtered)

	// filtered is false, all filters do not match
	fixtureAsset.Filters = []string{"entity.name == 'foo'", "entity.entity_class == 'host'"}
	filtered, err = filteredManager.isFiltered(fixtureAsset)
	assert.NoError(t, err)
	assert.False(t, filtered)

	// filtered is false, syntax error
	fixtureAsset.Filters = []string{"[(!@#$%^&"}
	filtered, err = filteredManager.isFiltered(fixtureAsset)
	assert.Error(t, err)
	assert.False(t, filtered)
}

func TestEvaluateAssetBuilds(t *testing.T) {
	_, _, filteredManager := NewTestFilteredManager()

	fixtureAsset := corev2.FixtureAsset("test-asset")
	fixtureAsset.Builds = []*corev2.AssetBuild{
		{
			URL: "asset-1",
			Filters: []string{
				"entity.name == 'asdf'",
			},
		},
		{
			URL: "asset-2",
			Filters: []string{
				"entity.name == 'test-entity'",
			},
		},
		{
			URL: "asset-3",
			Filters: []string{
				"entity.name == 'test-entity'",
				"entity.namespace == 'default'",
				"entity.system.arch == 'amd64'",
			},
		},
		{
			URL: "asset-4",
			Filters: []string{
				"entity.name == 'test-entity'",
				"entity.namespace == 'default'",
			},
		},
	}

	actualAsset, err := filteredManager.evaluateAssetBuilds(fixtureAsset)
	assert.NoError(t, err)
	assert.NotNil(t, actualAsset)
	assert.Equal(t, "asset-3", actualAsset.URL)
}
