package asset

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	fixturePath string
)

func init() {
	path, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	fixturePath = filepath.Join(path, "fixtures")
}

func getFixturePath(name string) string {
	return filepath.Join(fixturePath, name)
}

func TestFetchExistingAsset(t *testing.T) {
	t.Parallel()

	assetName := "rubby-on-rails.tar"
	localAssetPath := getFixturePath(assetName)

	fetcher := &httpFetcher{
		URLGetter: func(ctx context.Context, path string) (io.ReadCloser, error) {
			return os.Open(path)
		},
	}
	f, err := fetcher.Fetch(context.TODO(), localAssetPath)
	if err != nil {
		t.Logf("expected no error, got: %v", err)
		t.FailNow()
	}
	defer f.Close()
	defer os.Remove(f.Name())

	desiredSHA, _ := ioutil.ReadFile(getFixturePath(fmt.Sprintf("%s.sha512", assetName)))

	verifier := &sha512Verifier{}
	if err := verifier.Verify(f, string(desiredSHA)); err != nil {
		t.Logf("expected no error, got: %v", err)
		t.FailNow()
	}
}
