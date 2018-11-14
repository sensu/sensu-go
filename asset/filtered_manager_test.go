package asset

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type MockGetter struct {
	getCalled bool
	asset     *RuntimeAsset
	err       error
}

// Get satisfies the asset.Getter interface
func (m *MockGetter) Get(*types.Asset) (*RuntimeAsset, error) {
	m.getCalled = true
	return m.asset, m.err
}

// NewTestFilteredManager creates a new FilteredManager for testing
func NewTestFilteredManager() (*MockGetter, *types.Entity, *filteredManager) {
	mockGetter := &MockGetter{asset: &RuntimeAsset{Path: "/foo/bar"}}
	entity := types.FixtureEntity("test-entity")
	filteredManager := NewFilteredManager(mockGetter, entity)
	return mockGetter, entity, filteredManager
}

// FilteredManager should call the underlying Getter.
func TestFilteredManagerCallsGetter(t *testing.T) {
	mockGetter, _, filteredManager := NewTestFilteredManager()

	actualAsset, err := filteredManager.Get(types.FixtureAsset("test-asset"))
	assert.NoError(t, err)
	assert.Equal(t, mockGetter.asset, actualAsset)
	assert.True(t, mockGetter.getCalled)
}

// FilteredManager should call underlying Getter on filtered asset.
func TestFilteredManagerFilteredAsset(t *testing.T) {
	mockGetter, entity, filteredManager := NewTestFilteredManager()

	fixtureAsset := types.FixtureAsset("test-asset")
	fixtureAsset.Filters = []string{fmt.Sprintf("entity.name == '%s'", entity.Name)}
	actualAsset, err := filteredManager.Get(fixtureAsset)
	assert.NoError(t, err)
	assert.Equal(t, mockGetter.asset, actualAsset)
	assert.True(t, mockGetter.getCalled)
}

// FilteredManager should not call underlying Getter on unfiltered asset.
func TestFilteredManagerUnfilteredAsset(t *testing.T) {
	mockGetter, _, filteredManager := NewTestFilteredManager()

	fixtureAsset := types.FixtureAsset("test-asset")
	fixtureAsset.Filters = []string{"entity.name == 'foo'"}
	actualAsset, err := filteredManager.Get(fixtureAsset)
	assert.NoError(t, err)
	assert.Nil(t, actualAsset)
	assert.False(t, mockGetter.getCalled)
}

// FilteredManager should return error passed by underlying Getter.
func TestFilteredManagerError(t *testing.T) {
	mockGetter, _, filteredManager := NewTestFilteredManager()
	mockGetter.err = errors.New("TestFilteredManagerError")

	_, err := filteredManager.Get(types.FixtureAsset("test-asset"))
	assert.Error(t, err)
	assert.True(t, mockGetter.getCalled)
}

// isFiltered should allow filtering by entity
func TestIsFiltered(t *testing.T) {
	_, entity, filteredManager := NewTestFilteredManager()

	// filtered is true, filter matches
	fixtureAsset := types.FixtureAsset("test-asset")
	fixtureAsset.Filters = []string{fmt.Sprintf("entity.name == '%s'", entity.Name)}
	filtered, err := filteredManager.isFiltered(fixtureAsset)
	assert.NoError(t, err)
	assert.True(t, filtered)

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
