package bonsai

import (
	"reflect"
	"testing"

	goversion "github.com/hashicorp/go-version"
)

func TestAsset_BonsaiVersion(t *testing.T) {
	type fields struct {
		Name     string
		Versions []*AssetVersionGrouping
	}
	tests := []struct {
		name    string
		fields  fields
		version string
		want    string
		wantErr bool
	}{
		{
			name: "latest version is returned if none was specified",
			fields: fields{
				Name: "testasset",
				Versions: []*AssetVersionGrouping{
					&AssetVersionGrouping{Version: "0.1.0"},
					&AssetVersionGrouping{Version: "0.2.0"},
					&AssetVersionGrouping{Version: "v0.1.0"},
				},
			},
			want: "0.2.0",
		},
		{
			name: "exact requested version is returned",
			fields: fields{
				Name: "testasset",
				Versions: []*AssetVersionGrouping{
					&AssetVersionGrouping{Version: "0.1.0"},
					&AssetVersionGrouping{Version: "0.2.0"},
					&AssetVersionGrouping{Version: "v0.1.0"},
				},
			},
			version: "0.2.0",
			want:    "0.2.0",
		},
		{
			name: "requested version ignores the v prefix",
			fields: fields{
				Name: "testasset",
				Versions: []*AssetVersionGrouping{
					&AssetVersionGrouping{Version: "0.1.0"},
					&AssetVersionGrouping{Version: "0.2.0"},
					&AssetVersionGrouping{Version: "v0.1.0"},
				},
			},
			version: "v0.2.0",
			want:    "0.2.0",
		},
		{
			name: "requested version ignores the missing v prefix",
			fields: fields{
				Name: "testasset",
				Versions: []*AssetVersionGrouping{
					&AssetVersionGrouping{Version: "0.1.0"},
					&AssetVersionGrouping{Version: "v0.2.0"},
					&AssetVersionGrouping{Version: "v0.1.0"},
				},
			},
			version: "0.2.0",
			want:    "v0.2.0",
		},
		{
			name: "inexistant requested version returns an error",
			fields: fields{
				Name: "testasset",
				Versions: []*AssetVersionGrouping{
					&AssetVersionGrouping{Version: "0.1.0"},
					&AssetVersionGrouping{Version: "v0.2.0"},
					&AssetVersionGrouping{Version: "v0.1.0"},
				},
			},
			version: "0.3.0",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Asset{
				Name:     tt.fields.Name,
				Versions: tt.fields.Versions,
			}
			version, _ := goversion.NewVersion(tt.version)
			got, err := b.BonsaiVersion(version)
			if (err != nil) != tt.wantErr {
				t.Errorf("Asset.BonsaiVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want, _ := goversion.NewVersion(tt.want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("Asset.BonsaiVersion() = %v, want %v", got, want)
			}
		})
	}
}
