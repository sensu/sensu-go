package v2

import (
	"reflect"
	"testing"
)

func TestExtensionFields(t *testing.T) {
	tests := []struct {
		name    string
		args    Fielder
		wantKey string
		want    string
	}{
		{
			name:    "exposes name",
			args:    FixtureExtension("widget"),
			wantKey: "extension.name",
			want:    "widget",
		},
		{
			name: "exposes labels",
			args: &Extension{
				ObjectMeta: ObjectMeta{
					Labels: map[string]string{"region": "philadelphia"},
				},
			},
			wantKey: "extension.labels.region",
			want:    "philadelphia",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.Fields()
			if !reflect.DeepEqual(got[tt.wantKey], tt.want) {
				t.Errorf("Extension.Fields() = got[%s] %v, want[%s] %v", tt.wantKey, got[tt.wantKey], tt.wantKey, tt.want)
			}
		})
	}
}
