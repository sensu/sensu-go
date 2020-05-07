// Package asset provides a mechanism for installing, managing, and utilizing
// Sensu Assets.
//
// Access to assets are serialized. When an asset is first encountered,
// getting the asset from the manager blocks until the asset has been
// fetched, verified, and expanded on the host filesystem (or deemed
// unnecessary due to asset filters).
//
// The first goroutine to get an asset will cause the installation, and
// subsequent calling goroutines will simply block while installation
// completes. If the initial installation fails, the next goroutine to
// unblock will attempt reinstallation.
package asset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeAsset_Env(t *testing.T) {
	r := &RuntimeAsset{
		Name:   "foo",
		Path:   string(os.PathSeparator) + filepath.Join("tmp", "foo"),
		SHA512: "123456789",
	}
	got := r.Env()

	if want := string(os.PathSeparator) + filepath.Join("tmp", "foo", "bin"); !strings.Contains(got[0], want) {
		t.Errorf("RuntimeAsset.Env() PATH = %v, must contains %v", got[0], want)
	}
	if want := string(os.PathSeparator) + filepath.Join("tmp", "foo", "lib"); !strings.Contains(got[1], want) {
		t.Errorf("RuntimeAsset.Env() LD_LIBRARY_PATH = %v, must contains %v", got[1], want)
	}
	if want := string(os.PathSeparator) + filepath.Join("tmp", "foo", "include"); !strings.Contains(got[2], want) {
		t.Errorf("RuntimeAsset.Env() CPATH = %v, must contains %v", got[2], want)
	}
	if want := string(os.PathSeparator) + filepath.Join("tmp", "foo"); !strings.Contains(got[3], want) {
		t.Errorf("RuntimeAsset.Env() asset path = %v, must contains %v", got[3], want)
	}
}
