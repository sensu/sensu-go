package asset

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sensu/sensu-go/asset"
)

func TestExpandValidTar(t *testing.T) {
	t.Parallel()
	assetPath := getFixturePath("rubby-on-rails.tar")
	f, err := os.Open(assetPath)
	if err != nil {
		t.Fatalf("unable to open asset fixture, err: %v", err)
	}
	defer f.Close()

	tmpDir := os.TempDir()
	targetDirectory := filepath.Join(tmpDir, "rubby-on-rails-tar")
	if err := os.Mkdir(targetDirectory, 0755); err != nil {
		t.Fatalf("unable to create target directory, err: %v", err)
	}
	defer os.RemoveAll(targetDirectory)

	expander := &asset.archiveExpander{}
	if err := expander.Expand(f, targetDirectory); err != nil {
		t.Logf("expected no error, got %v", err)
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

	tmpDir := os.TempDir()
	targetDirectory := filepath.Join(tmpDir, "rubby-on-rails-tgz")
	if err := os.Mkdir(targetDirectory, 0755); err != nil {
		t.Fatalf("unable to create target directory, err: %v", err)
	}
	defer os.RemoveAll(targetDirectory)

	expander := &asset.archiveExpander{}
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

	tmpDir := os.TempDir()
	targetDirectory := filepath.Join(tmpDir, "unsupported-zip")
	if err := os.Mkdir(targetDirectory, 0755); err != nil {
		t.Fatalf("unable to create target directory, err: %v", err)
	}
	defer os.RemoveAll(targetDirectory)

	expander := &asset.archiveExpander{}
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

	tmpDir := os.TempDir()
	targetDirectory := filepath.Join(tmpDir, "invalid-tar")
	if err := os.Mkdir(targetDirectory, 0755); err != nil {
		t.Fatalf("unable to create target directory, err: %v", err)
	}
	defer os.RemoveAll(targetDirectory)

	expander := &asset.archiveExpander{}
	if err := expander.Expand(f, targetDirectory); err == nil {
		t.Log("expected error, got nil")
		t.Fail()
	}
}
