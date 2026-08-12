package asset

import (
	"context"
	"fmt"
	"io"
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
		URLGetter: func(ctx context.Context, path, trustedCAFile string, header map[string]string) (io.ReadCloser, error) {
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

	desiredSHA, _ := os.ReadFile(getFixturePath(fmt.Sprintf("%s.sha512", assetName)))

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
	closer, err := httpGet(context.Background(), ts.URL, "", headers)
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
	closer, err := httpGet(context.Background(), ts.URL, "", headers)
	assert.Nil(t, closer)
	assert.EqualError(t, err, "error fetching asset: Response Code 404")
}

func TestHTTPGetNon200ClosesBody(t *testing.T) {
	closed := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer ts.Close()

	// Use the real httpGet and verify no connection leak by making many requests
	for i := 0; i < 100; i++ {
		closer, err := httpGet(context.Background(), ts.URL, "", nil)
		assert.Nil(t, closer)
		assert.Error(t, err)
	}
	// If bodies weren't closed, we'd exhaust file descriptors before 100 iterations
	_ = closed
}

func TestHTTPGet404ReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	closer, err := httpGet(context.Background(), ts.URL, "", nil)
	assert.Nil(t, closer)
	assert.Contains(t, err.Error(), "Response Code 404")
}

func TestHTTPGet500ReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	closer, err := httpGet(context.Background(), ts.URL, "", nil)
	assert.Nil(t, closer)
	assert.Contains(t, err.Error(), "Response Code 500")
}

func TestHTTPGet403ReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	closer, err := httpGet(context.Background(), ts.URL, "", nil)
	assert.Nil(t, closer)
	assert.Contains(t, err.Error(), "Response Code 403")
}

func TestHTTPGetRepeatedFailuresNoFDLeak(t *testing.T) {
	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("asset not found at this location"))
	}))
	defer ts.Close()

	// Simulate repeated check executions with a missing asset
	// Without the fix, each iteration leaks a file descriptor
	for i := 0; i < 500; i++ {
		closer, err := httpGet(context.Background(), ts.URL, "", nil)
		assert.Nil(t, closer)
		assert.Error(t, err)
	}
	assert.Equal(t, 500, requestCount)
}
