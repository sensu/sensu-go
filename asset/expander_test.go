package asset

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sensu/sensu-go/testing/testutil"
)

func TestExpandValidTar(t *testing.T) {
	t.Parallel()
	assetPath := getFixturePath("rubby-on-rails.tar")
	f, err := os.Open(assetPath)
	if err != nil {
		t.Fatalf("unable to open asset fixture, err: %v", err)
	}
	defer f.Close()

	tmpDir, remove := testutil.TempDir(t)
	defer remove()
	targetDirectory := filepath.Join(tmpDir, "rubby-on-rails-tar")
	if err := os.Mkdir(targetDirectory, 0755); err != nil {
		t.Fatalf("unable to create target directory, err: %v", err)
	}

	expander := &archiveExpander{}
	if err := expander.Expand(f, targetDirectory); err != nil {
		t.Logf("expected no error, got %v", err)
		t.Fail()
	}

	railsFile, err := os.Open(filepath.Join(targetDirectory, "bin", "rails"))
	if err != nil {
		t.Logf("could not open asset contents, err: %v", err)
		t.Fail()
	}

	if railsFile == nil {
		t.Logf("no file opened")
		t.Fail()
	}
}

func TestExpandValidTGZ(t *testing.T) {
	t.Parallel()

	assetPath := getFixturePath("rubby-on-rails.tgz")
	f, err := os.Open(assetPath)
	if err != nil {
		t.Fatalf("unable to open asset fixture, err: %v", err)
	}
	defer f.Close()

	tmpDir, remove := testutil.TempDir(t)
	defer remove()
	targetDirectory := filepath.Join(tmpDir, "rubby-on-rails-tgz")
	if err := os.Mkdir(targetDirectory, 0755); err != nil {
		t.Fatalf("unable to create target directory, err: %v", err)
	}

	expander := &archiveExpander{}
	if err := expander.Expand(f, targetDirectory); err != nil {
		t.Logf("expected no error, got %v", err)
		t.Fail()
	}
}

func TestExpandUnsupportedArchive(t *testing.T) {
	t.Parallel()

	assetPath := getFixturePath("unsupported.zip")
	f, err := os.Open(assetPath)
	if err != nil {
		t.Fatalf("unable to open asset fixture, err: %v", err)
	}
	defer f.Close()

	tmpDir, remove := testutil.TempDir(t)
	defer remove()
	targetDirectory := filepath.Join(tmpDir, "unsupported-zip")
	if err := os.Mkdir(targetDirectory, 0755); err != nil {
		t.Fatalf("unable to create target directory, err: %v", err)
	}

	expander := &archiveExpander{}
	if err := expander.Expand(f, targetDirectory); err == nil {
		t.Log("expected error, got nil")
		t.Fail()
	}
}

func TestExpandInvalidArchive(t *testing.T) {
	t.Parallel()

	assetPath := getFixturePath("invalid.tar")
	f, err := os.Open(assetPath)
	if err != nil {
		t.Fatalf("unable to open asset fixture, err: %v", err)
	}
	defer f.Close()

	tmpDir, remove := testutil.TempDir(t)
	defer remove()
	targetDirectory := filepath.Join(tmpDir, "invalid-tar")
	if err := os.Mkdir(targetDirectory, 0755); err != nil {
		t.Fatalf("unable to create target directory, err: %v", err)
	}

	expander := &archiveExpander{}
	if err := expander.Expand(f, targetDirectory); err == nil {
		t.Log("expected error, got nil")
		t.Fail()
	}
}
