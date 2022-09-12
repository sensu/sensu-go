package v3

import (
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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
			args:    &EntityConfig{Metadata: &corev2.ObjectMeta{}, Deregister: true},
			wantKey: "entity_config.deregister",
			want:    "true",
		},
		{
			name:    "exposes class",
			args:    &EntityConfig{Metadata: &corev2.ObjectMeta{}, EntityClass: "agent"},
			wantKey: "entity_config.entity_class",
			want:    "agent",
		},
		{
			name:    "exposes subscriptions",
			args:    &EntityConfig{Metadata: &corev2.ObjectMeta{}, Subscriptions: []string{"www", "unix"}},
			wantKey: "entity_config.subscriptions",
			want:    "www,unix",
		},
		{
			name: "exposes labels",
			args: &EntityConfig{
				Metadata: &corev2.ObjectMeta{
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

func TestEntityConfig_ProduceRedacted(t *testing.T) {
	tests := []struct {
		name string
		in   *EntityConfig
		want *EntityConfig
	}{
		{
			name: "nil metadata",
			in: func() *EntityConfig {
				cfg := FixtureEntityConfig("test")
				cfg.Metadata = nil
				return cfg
			}(),
			want: func() *EntityConfig {
				cfg := FixtureEntityConfig("test")
				cfg.Metadata = nil
				return cfg
			}(),
		},
		{
			name: "nothing to redact",
			in: func() *EntityConfig {
				cfg := FixtureEntityConfig("test")
				cfg.Metadata.Labels["my_field"] = "test123"
				return cfg
			}(),
			want: func() *EntityConfig {
				cfg := FixtureEntityConfig("test")
				cfg.Metadata.Labels["my_field"] = "test123"
				return cfg
			}(),
		},
		{
			name: "redact default fields",
			in: func() *EntityConfig {
				cfg := FixtureEntityConfig("test")
				cfg.Metadata.Labels["my_field"] = "test123"
				cfg.Metadata.Labels["password"] = "test123"
				return cfg
			}(),
			want: func() *EntityConfig {
				cfg := FixtureEntityConfig("test")
				cfg.Metadata.Labels["my_field"] = "test123"
				cfg.Metadata.Labels["password"] = corev2.Redacted
				return cfg
			}(),
		},
		{
			name: "redact custom fields",
			in: func() *EntityConfig {
				cfg := FixtureEntityConfig("test")
				cfg.Redact = []string{"my_field"}
				cfg.Metadata.Labels["my_field"] = "test123"
				cfg.Metadata.Labels["password"] = "test123"
				return cfg
			}(),
			want: func() *EntityConfig {
				cfg := FixtureEntityConfig("test")
				cfg.Redact = []string{"my_field"}
				cfg.Metadata.Labels["my_field"] = corev2.Redacted
				cfg.Metadata.Labels["password"] = "test123"
				return cfg
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.in
			if got := e.ProduceRedacted(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EntityConfig.ProduceRedacted() = %v, want %v", got, tt.want)
			}
		})
	}
}
