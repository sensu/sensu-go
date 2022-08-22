package v3

import (
	"reflect"
	"testing"
)

func TestEntityStateFields(t *testing.T) {
	tests := []struct {
		name    string
		args    Fielder
		wantKey string
		want    string
	}{
		{
			name:    "exposes name",
			args:    FixtureEntityState("my-agent"),
			wantKey: "entity_state.name",
			want:    "my-agent",
		},
		{
			name:    "exposes deregister",
			args:    FixtureEntityState("my-agent"),
			wantKey: "entity_state.namespace",
			want:    "default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.Fields()
			if !reflect.DeepEqual(got[tt.wantKey], tt.want) {
				t.Errorf("EntityState.Fields() = got[%s] %v, want[%s] %v", tt.wantKey, got[tt.wantKey], tt.wantKey, tt.want)
			}
		})
	}
}
