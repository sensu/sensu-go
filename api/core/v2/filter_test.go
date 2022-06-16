package v2

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureEventFilter(t *testing.T) {
	filter := FixtureEventFilter("filter")
	assert.Equal(t, "filter", filter.Name)
	assert.Equal(t, EventFilterActionAllow, filter.Action)
	assert.Equal(t, []string{"event.check.team == 'ops'"}, filter.Expressions)
	assert.NoError(t, filter.Validate())
}

func TestFixtureDenyEventFilter(t *testing.T) {
	filter := FixtureDenyEventFilter("filter")
	assert.Equal(t, "filter", filter.Name)
	assert.Equal(t, EventFilterActionDeny, filter.Action)
	assert.Equal(t, []string{"event.check.team == 'ops'"}, filter.Expressions)
	assert.NoError(t, filter.Validate())
}

func TestEventFilterValidate(t *testing.T) {
	var f EventFilter

	// Invalid name
	assert.Error(t, f.Validate())
	f.Name = "foo"

	// Invalid action
	assert.Error(t, f.Validate())
	f.Action = "allow"

	// Invalid attributes
	assert.Error(t, f.Validate())
	f.Expressions = []string{"event.check.team == 'ops'"}

	// Invalid namespace
	assert.Error(t, f.Validate())
	f.Namespace = "default"

	// Valid filter
	assert.NoError(t, f.Validate())
}

func TestEventFilterFields(t *testing.T) {
	tests := []struct {
		name    string
		args    Fielder
		wantKey string
		want    string
	}{
		{
			name:    "exposes name",
			args:    FixtureEventFilter("allow-silenced"),
			wantKey: "filter.name",
			want:    "allow-silenced",
		},
		{
			name:    "exposes action",
			args:    &EventFilter{Action: "allow"},
			wantKey: "filter.action",
			want:    "allow",
		},
		{
			name: "exposes labels",
			args: &EventFilter{
				ObjectMeta: ObjectMeta{
					Labels: map[string]string{"region": "philadelphia"},
				},
			},
			wantKey: "filter.labels.region",
			want:    "philadelphia",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.Fields()
			if !reflect.DeepEqual(got[tt.wantKey], tt.want) {
				t.Errorf("EventFilter.Fields() = got[%s] %v, want[%s] %v", tt.wantKey, got[tt.wantKey], tt.wantKey, tt.want)
			}
		})
	}
}
