package v3

import (
	"reflect"
	"testing"

	v2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestEntityConfigFields(t *testing.T) {
	tests := []struct {
		name    string
		args    Fielder
		wantKey string
		want    string
	}{
		{
			name:    "exposes name",
			args:    FixtureEntityConfig("my-agent"),
			wantKey: "entity_config.name",
			want:    "my-agent",
		},
		{
			name:    "exposes deregister",
			args:    &EntityConfig{Metadata: &v2.ObjectMeta{}, Deregister: true},
			wantKey: "entity_config.deregister",
			want:    "true",
		},
		{
			name:    "exposes class",
			args:    &EntityConfig{Metadata: &v2.ObjectMeta{}, EntityClass: "agent"},
			wantKey: "entity_config.entity_class",
			want:    "agent",
		},
		{
			name:    "exposes subscriptions",
			args:    &EntityConfig{Metadata: &v2.ObjectMeta{}, Subscriptions: []string{"www", "unix"}},
			wantKey: "entity_config.subscriptions",
			want:    "www,unix",
		},
		{
			name: "exposes labels",
			args: &EntityConfig{
				Metadata: &v2.ObjectMeta{
					Labels: map[string]string{"region": "philadelphia"},
				},
			},
			wantKey: "entity_config.labels.region",
			want:    "philadelphia",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.Fields()
			if !reflect.DeepEqual(got[tt.wantKey], tt.want) {
				t.Errorf("EntityConfig.Fields() = got[%s] %v, want[%s] %v", tt.wantKey, got[tt.wantKey], tt.wantKey, tt.want)
			}
		})
	}
}
