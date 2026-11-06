package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReapRemovesOldArchives(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	old := filepath.Join(dir, fmt.Sprintf("app.log.%d", time.Now().Add(-48*time.Hour).UnixNano()))
	recent := filepath.Join(dir, fmt.Sprintf("app.log.%d", time.Now().UnixNano()))
	for _, f := range []string{old, recent} {
		if err := os.WriteFile(f, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	r := &RotateFileLogger{path: path, retentionDuration: time.Hour}
	errs := make(chan error, 4)
	r.reap(errs)

	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Errorf("archive older than the retention duration was not removed: %v", err)
	}
	if _, err := os.Stat(recent); err != nil {
		t.Errorf("recent archive was removed: %v", err)
	}
}
