package asset

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openFDs returns the number of open file descriptors for this process, or
// skips the test on platforms that do not expose a listing.
func openFDs(t *testing.T) int {
	t.Helper()
	for _, dir := range []string{"/proc/self/fd", "/dev/fd"} {
		if entries, err := os.ReadDir(dir); err == nil {
			return len(entries)
		}
	}
	t.Skip("file descriptor listing not available")
	return 0
}

func newLibDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "lib", "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "lib", "a.js"), []byte("var a=1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "lib", "sub", "b.js"), []byte("var b=1"), 0644))
	return dir
}

func TestScriptsReturnsAllJSFiles(t *testing.T) {
	set := RuntimeAssetSet{&RuntimeAsset{Path: newLibDir(t)}}
	scripts, err := set.Scripts()
	require.NoError(t, err)
	defer func() {
		for _, f := range scripts {
			f.Close()
		}
	}()
	assert.Len(t, scripts, 2)
}

// TestScriptsErrorsOnUnreadableDir checks that an unreadable subdirectory is an
// error, not a silent skip, and that files already opened before the failure
// are closed.
func TestScriptsErrorsOnUnreadableDir(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("permission checks are not enforced for this user or platform")
	}
	dir := newLibDir(t)
	locked := filepath.Join(dir, "lib", "zz-locked")
	require.NoError(t, os.MkdirAll(locked, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(locked, "c.js"), []byte("x"), 0644))
	require.NoError(t, os.Chmod(locked, 0000))
	t.Cleanup(func() { os.Chmod(locked, 0755) })

	set := RuntimeAssetSet{&RuntimeAsset{Path: dir}}
	before := openFDs(t)
	for i := 0; i < 50; i++ {
		scripts, err := set.Scripts()
		assert.Error(t, err, "unreadable directory must be reported")
		assert.Nil(t, scripts)
	}
	assert.LessOrEqual(t, openFDs(t)-before, 5, "failed walks must not leak file descriptors")
}
