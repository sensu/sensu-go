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
