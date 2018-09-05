package asset_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/sensu/sensu-go/asset"
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

type mockURLGetter struct {
	filePath string
}

func (m *mockURLGetter) Get(string) (io.ReadCloser, error) {
	return os.Open(m.filePath)
}

func TestFetchExistingAsset(t *testing.T) {
	t.Parallel()

	assetName := "rubby-on-rails.tar"

	fetcher := &asset.HTTPFetcher{}
	fetcher.InjectGetter(&mockURLGetter{getFixturePath(assetName)})
	f, err := fetcher.Fetch("")
	if err != nil {
		t.Logf("expected no error, got: %v", err)
		t.FailNow()
	}
	defer f.Close()
	defer os.Remove(f.Name())

	desiredSHA, _ := ioutil.ReadFile(getFixturePath(fmt.Sprintf("%s.sha512", assetName)))

	verifier := &asset.SHA512Verifier{}
	if err := verifier.Verify(f, string(desiredSHA)); err != nil {
		t.Logf("expected no error, got: %v", err)
		t.FailNow()
	}
}
