package asset

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
	var headers map[string]string

	fetcher := &httpFetcher{
		URLGetter: func(ctx context.Context, path string, header map[string]string) (io.ReadCloser, error) {
			return os.Open(path)
		},
	}
	f, err := fetcher.Fetch(context.TODO(), localAssetPath, headers)
	if err != nil {
		t.Logf("expected no error, got: %v", err)
		t.FailNow()
	}
	defer f.Close()
	defer os.Remove(f.Name())

	desiredSHA, _ := ioutil.ReadFile(getFixturePath(fmt.Sprintf("%s.sha512", assetName)))

	verifier := &Sha512Verifier{}
	if err := verifier.Verify(f, string(desiredSHA)); err != nil {
		t.Logf("expected no error, got: %v", err)
		t.FailNow()
	}
}

func TestHTTPGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, []string{"bar"}, r.Header["Foo"])
		assert.Equal(t, []string{"hello", "sup"}, r.Header["Hi"])
	}))
	defer ts.Close()

	headers := map[string]string{"foo": "bar", "hi": "hello, sup"}
	closer, err := httpGet(context.Background(), ts.URL, headers)
	assert.NotNil(t, closer)
	assert.NoError(t, err)
}

func TestHTTPGetNon200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, []string{"bar"}, r.Header["Foo"])
		assert.Equal(t, []string{"hello", "sup"}, r.Header["Hi"])
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	headers := map[string]string{"foo": "bar", "hi": "hello, sup"}
	closer, err := httpGet(context.Background(), ts.URL, headers)
	assert.Nil(t, closer)
	assert.EqualError(t, err, "error fetching asset: Response Code 404")
}
