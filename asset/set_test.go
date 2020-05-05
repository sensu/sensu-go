package asset

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func fixtureAssets() []types.Asset {
	return []types.Asset{
		*types.FixtureAsset("asset-1"),
		*types.FixtureAsset("asset-2"),
		*types.FixtureAsset("asset-3"),
	}
}

func fixtureRuntimeAssets() RuntimeAssetSet {
	return RuntimeAssetSet{
		&RuntimeAsset{Name: "foo1", Path: string(os.PathSeparator) + filepath.Join("foo", "bar", "asset-1")},
		&RuntimeAsset{Name: "foo2", Path: string(os.PathSeparator) + filepath.Join("foo", "bar", "asset-2")},
		&RuntimeAsset{Name: "foo3", Path: string(os.PathSeparator) + filepath.Join("foo", "bar", "asset-3")},
	}
}

type mockGetter struct {
	err error
}

func (m *mockGetter) Get(ctx context.Context, asset *corev2.Asset) (*RuntimeAsset, error) {
	if m.err != nil {
		return nil, m.err
	}
	var runtimeAsset *RuntimeAsset
	if asset.Name == "nil" {
		runtimeAsset = nil
	} else {
		runtimeAsset = &RuntimeAsset{Path: fmt.Sprintf("/foo/bar/%s", asset.Name)}
	}
	return runtimeAsset, nil
}

// GetAll should output a RuntimeAssetSet based on input Assets
func TestGetAll(t *testing.T) {
	assets := fixtureAssets()
	mockGetter := mockGetter{}
	expectedRuntimeAssets := RuntimeAssetSet{}
	for _, asset := range assets {
		runtimeAsset, _ := mockGetter.Get(context.TODO(), &asset)
		expectedRuntimeAssets = append(expectedRuntimeAssets, runtimeAsset)
	}

	actualRuntimeAssets, err := GetAll(context.TODO(), &mockGetter, assets)
	assert.NoError(t, err)
	assert.Equal(t, expectedRuntimeAssets, actualRuntimeAssets)
}

// GetAll should return first error encountered
func TestGetAllError(t *testing.T) {
	assets := fixtureAssets()
	mockGetter := mockGetter{err: errors.New("test error")}
	expectedRuntimeAssets := (RuntimeAssetSet)(nil)

	actualRuntimeAssets, err := GetAll(context.TODO(), &mockGetter, assets)
	assert.Error(t, err)
	assert.Equal(t, expectedRuntimeAssets, actualRuntimeAssets)
}

// GetAll should filter nil values
func TestGetAllFilterNil(t *testing.T) {
	assets := fixtureAssets()
	assets[1].Name = "nil"
	mockGetter := mockGetter{}

	actualRuntimeAssets, err := GetAll(context.TODO(), &mockGetter, assets)
	assert.NoError(t, err)
	assert.NotContains(t, actualRuntimeAssets, nil)
}

// Env should contain paths for each asset in RuntimeAssetSet
func TestEnvContainsPaths(t *testing.T) {
	runtimeAssetSet := fixtureRuntimeAssets()
	env := runtimeAssetSet.Env()
	for _, envVar := range env {
		for _, runtimeAsset := range runtimeAssetSet {
			if strings.HasPrefix(envVar, "FOO") {
				continue
			}
			assert.Contains(t, envVar, runtimeAsset.Path)
		}
	}
}

// Env should contain existing environment values if applicable
func TestEnvContainsParentEnv(t *testing.T) {
	envKey := "PATH"
	runtimeAssetSet := fixtureRuntimeAssets()
	oldEnv := os.Getenv(envKey)
	testPath := "/parent/env/path"
	os.Setenv(envKey, testPath)
	env := runtimeAssetSet.Env()

	keyFound := false
	for _, envVar := range env {
		key := strings.Split(envVar, "=")[0]
		if key == envKey {
			keyFound = true
			assert.Contains(t, envVar, testPath)
		}
	}
	assert.True(t, keyFound)
	os.Setenv(envKey, oldEnv)
}
