package v2

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureAPIKey(t *testing.T) {
	a := FixtureAPIKey("226f9e06-9d54-45c6-a9f6-4206bfa7ccf6", "bar")
	assert.NoError(t, a.Validate())
	assert.Equal(t, "226f9e06-9d54-45c6-a9f6-4206bfa7ccf6", a.Name)
	assert.Equal(t, "bar", a.Username)
	assert.Equal(t, "", a.Namespace)
}

func TestAPIKeyValidate(t *testing.T) {
	a := &APIKey{}

	// Namespace
	a.Namespace = "foo"
	assert.Error(t, a.Validate())
	a.Namespace = ""

	// Empty username
	assert.Error(t, a.Validate())
	a.Username = "bar"

	// Empty name
	assert.Error(t, a.Validate())
	a.Name = "foo"

	// Invalid name
	assert.Error(t, a.Validate())
	a.Name = "226f9e06-9d54-45c6-a9f6-4206bfa7ccf6"

	assert.NoError(t, a.Validate())
	assert.Equal(t, "226f9e06-9d54-45c6-a9f6-4206bfa7ccf6", a.Name)
	assert.Equal(t, "bar", a.Username)
	assert.Equal(t, "", a.Namespace)
}

func TestAPIKeyFields(t *testing.T) {
	tests := []struct {
		name    string
		args    Resource
		wantKey string
		want    string
	}{
		{
			name:    "exposes name",
			args:    FixtureAPIKey("circle-ci-access", "admin"),
			wantKey: "api_key.name",
			want:    "circle-ci-access",
		},
		{
			name:    "exposes username",
			args:    FixtureAPIKey("circle-ci-access", "admin"),
			wantKey: "api_key.username",
			want:    "admin",
		},
		{
			name: "exposes labels",
			args: &APIKey{
				ObjectMeta: ObjectMeta{
					Labels: map[string]string{"region": "philadelphia"},
				},
			},
			wantKey: "api_key.labels.region",
			want:    "philadelphia",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := APIKeyFields(tt.args)
			if !reflect.DeepEqual(got[tt.wantKey], tt.want) {
				t.Errorf("APIKeyFields() = got[%s] %v, want[%s] %v", tt.wantKey, got[tt.wantKey], tt.wantKey, tt.want)
			}
		})
	}
}
