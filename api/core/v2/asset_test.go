package v2

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureCreatesValidAsset(t *testing.T) {
	assert := assert.New(t)
	a := FixtureAsset("one")
	assert.Equal("one", a.Name)
	assert.NoError(a.Validate())
}

func TestValidator(t *testing.T) {
	assert := assert.New(t)
	asset := FixtureAsset("one")

	// Given valid asset it should pass
	assert.NoError(asset.Validate())

	// Given asset without a name it should not pass
	asset = FixtureAsset("")
	assert.Error(asset.Validate())

	// Given asset without a URL it should not pass
	asset = FixtureAsset("name")
	asset.URL = ""
	assert.Error(asset.Validate())

	// Given asset without an namespace it should not pass
	asset = FixtureAsset("name")
	asset.Namespace = ""
	assert.Error(asset.Validate())

	// Given asset with valid filters
	asset = FixtureAsset("name")
	asset.Filters = []string{`entity.OS in ("macos", "linux")`}
	assert.NoError(asset.Validate())

	// Given asset without a Sha512 it should not pass
	asset = FixtureAsset("name")
	asset.Sha512 = ""
	assert.Error(asset.Validate())

	// Given asset with an invalid Sha512 it should not pass
	asset = FixtureAsset("name")
	asset.Sha512 = "nope"
	assert.Error(asset.Validate())

	// Bonsai assets with uppercases should pass
	asset = FixtureAsset("Username/asset_name:0.0.1")
	assert.NoError(asset.Validate())
}

func TestValidateName_GH3344(t *testing.T) {
	assert := assert.New(t)
	asset := FixtureAsset("my-asset:1.0.2")
	assert.NoError(asset.Validate())
}

func TestAssetFields(t *testing.T) {
	tests := []struct {
		name    string
		args    Resource
		wantKey string
		want    string
	}{
		{
			name:    "exposes name",
			args:    FixtureAsset("curl"),
			wantKey: "asset.name",
			want:    "curl",
		},
		{
			name:    "exposes filters",
			args:    &Asset{Filters: []string{"test"}},
			wantKey: "asset.filters",
			want:    "test",
		},
		{
			name: "exposes labels",
			args: &Asset{
				ObjectMeta: ObjectMeta{
					Labels: map[string]string{"region": "philadelphia"},
				},
			},
			wantKey: "asset.labels.region",
			want:    "philadelphia",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AssetFields(tt.args)
			if !reflect.DeepEqual(got[tt.wantKey], tt.want) {
				t.Errorf("AssetFields() = got[%s] %v, want[%s] %v", tt.wantKey, got[tt.wantKey], tt.wantKey, tt.want)
			}
		})
	}
}
