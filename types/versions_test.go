package types_test

import (
	"testing"

	"runtime/debug"

	types "github.com/sensu/sensu-go/types"
)

// TODO(eric): This test doesn't work yet because of https://github.com/golang/go/issues/33976
func TestAPIModuleVersions(t *testing.T) {
	_, ok := debug.ReadBuildInfo()
	if ok {
		t.Fatal("remove this if block, the test works now")
	} else {
		t.Skip()
	}
	modVersions := types.APIModuleVersions()
	if _, ok := modVersions["core/v2"]; !ok {
		t.Errorf("missing core/v2 module version")
	}
	if _, ok := modVersions["core/v3"]; !ok {
		t.Errorf("missing core/v3 module version")
	}
}

func TestParseAPIVersion(t *testing.T) {
	tests := []struct {
		Input       string
		ExpAPIGroup string
		ExpSemVer   string
	}{
		{
			Input:       "core/v2",
			ExpAPIGroup: "core/v2",
			ExpSemVer:   "v2.0.0",
		},
		{
			Input:       "core/v2.1",
			ExpAPIGroup: "core/v2",
			ExpSemVer:   "v2.1.0",
		},
		{
			Input:       "core/v2.1.2",
			ExpAPIGroup: "core/v2",
			ExpSemVer:   "v2.1.2",
		},
		{
			Input:       "corev2.0.0",
			ExpAPIGroup: "corev2.0.0",
			ExpSemVer:   "v0.0.0",
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			group, version := types.ParseAPIVersion(test.Input)
			if got, want := group, test.ExpAPIGroup; got != want {
				t.Errorf("bad API group: got %q, want %q", got, want)
			}
			if got, want := version, test.ExpSemVer; got != want {
				t.Errorf("bad semver: got %q, want %q", got, want)
			}
		})
	}
}
