package v2

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureUser(t *testing.T) {
	u := FixtureUser("foo")
	assert.NoError(t, u.Validate())
	assert.Equal(t, "foo", u.Username)
	assert.Contains(t, u.Groups, "default")
}

func TestUserValidate(t *testing.T) {
	u := &User{}

	// Empty username
	assert.Error(t, u.Validate())

	u = FixtureUser("foo")
	assert.Equal(t, "foo", u.Username)
	assert.NoError(t, u.Validate())
}

func TestUserValidatePassword(t *testing.T) {
	u := &User{}

	// Empty password
	assert.Error(t, u.ValidatePassword())

	// Too short password
	u = FixtureUser("foo")
	u.Password = "123"
	assert.Error(t, u.ValidatePassword())

	u.Password = "P@ssw0rd!"
	assert.NoError(t, u.ValidatePassword())
}

func TestUserFields(t *testing.T) {
	tests := []struct {
		name    string
		args    Resource
		wantKey string
		want    string
	}{
		{
			name:    "exposes username",
			args:    FixtureUser("frank"),
			wantKey: "user.username",
			want:    "frank",
		},
		{
			name:    "exposes disabled",
			args:    FixtureUser("frank"),
			wantKey: "user.disabled",
			want:    "false",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserFields(tt.args)
			if !reflect.DeepEqual(got[tt.wantKey], tt.want) {
				t.Errorf("UserFields() = got[%s] %v, want[%s] %v", tt.wantKey, got[tt.wantKey], tt.wantKey, tt.want)
			}
		})
	}
}
