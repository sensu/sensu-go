package graphql

import (
	"reflect"
	"testing"
	"time"

	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/version"
)

func Test_versionsImpl_Backend(t *testing.T) {
	tests := []struct {
		name    string
		want    interface{}
		wantErr bool
	}{
		{
			name:    "main",
			want:    struct{}{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &versionsImpl{}
			got, err := r.Backend(graphql.ResolveParams{})
			if (err != nil) != tt.wantErr {
				t.Errorf("versionsImpl.Backend() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("versionsImpl.Backend() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensuBackendVersionImpl_Version(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "main",
			want:    version.Semver(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &sensuBackendVersionImpl{}
			got, err := r.Version(graphql.ResolveParams{})
			if (err != nil) != tt.wantErr {
				t.Errorf("sensuBackendVersionImpl.Version() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sensuBackendVersionImpl.Version() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensuBackendVersionImpl_BuildSHA(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{
			name:    "main",
			want:    version.BuildSHA,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &sensuBackendVersionImpl{}
			got, err := r.BuildSHA(graphql.ResolveParams{})
			if (err != nil) != tt.wantErr {
				t.Errorf("sensuBackendVersionImpl.BuildSHA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sensuBackendVersionImpl.BuildSHA() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sensuBackendVersionImpl_BuildDate(t *testing.T) {
	testT, _ := time.Parse(time.RFC3339, "2019-11-19T20:18:33Z")

	tests := []struct {
		name    string
		setup   func()
		want    *time.Time
		wantErr bool
	}{
		{
			name: "filled",
			setup: func() {
				version.BuildDate = "2019-11-19T20:18:33Z"
			},
			want:    &testT,
			wantErr: false,
		},
		{
			name: "empty",
			setup: func() {
				version.BuildDate = ""
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevT := version.BuildDate
			defer func() { version.BuildDate = prevT }()
			tt.setup()

			r := &sensuBackendVersionImpl{}
			got, err := r.BuildDate(graphql.ResolveParams{})
			if (err != nil) != tt.wantErr {
				t.Errorf("sensuBackendVersionImpl.BuildDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sensuBackendVersionImpl.BuildDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
