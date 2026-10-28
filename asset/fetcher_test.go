package asset

import (
	"context"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

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

// openConnTracker counts server-side connections that are open, using the
// server's ConnState hook. Connections that the client never closes stay open
// here, so the count shows a leak without relying on FD or goroutine counts.
type openConnTracker struct {
	mu   sync.Mutex
	open int
}

func (c *openConnTracker) hook(_ net.Conn, state http.ConnState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	switch state {
	case http.StateNew:
		c.open++
	case http.StateClosed, http.StateHijacked:
		c.open--
	}
}

func (c *openConnTracker) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.open
}

// TestHTTPGetBoundedConnections verifies that repeated asset fetches reuse
// connections instead of opening a new transport per call. Before the fix,
// every request through a CA-file client left its connection open, so 200
// requests produced about 200 open connections.
func TestHTTPGetBoundedConnections(t *testing.T) {
	const requests = 200
	const maxOpenConns = 5

	tests := []struct {
		name   string
		useTLS bool
		status int
		useCA  bool
	}{
		{name: "HTTP 404 no CA", useTLS: false, status: http.StatusNotFound, useCA: false},
		{name: "TLS 404 with CA", useTLS: true, status: http.StatusNotFound, useCA: true},
		{name: "TLS 200 with CA", useTLS: true, status: http.StatusOK, useCA: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tracker := &openConnTracker{}
			handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				w.Write([]byte("body"))
			})
			ts := httptest.NewUnstartedServer(handler)
			ts.Config.ConnState = tracker.hook
			if tc.useTLS {
				ts.StartTLS()
			} else {
				ts.Start()
			}
			defer ts.Close()

			caFile := ""
			if tc.useCA {
				caFile = writeServerCA(t, ts)
			}

			for i := 0; i < requests; i++ {
				closer, err := httpGet(context.Background(), ts.URL, caFile, nil)
				if err == nil {
					closer.Close()
				}
			}

			// Allow the server to observe closes from the client's idle timeout path.
			time.Sleep(100 * time.Millisecond)
			assert.LessOrEqual(t, tracker.count(), maxOpenConns,
				"open server connections after %d requests", requests)
		})
	}
}

// writeServerCA writes the test server's certificate to a PEM file and returns
// its path, so the fetcher can trust it through trustedCAFile.
func writeServerCA(t *testing.T, ts *httptest.Server) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ca.pem")
	block := &pem.Block{Type: "CERTIFICATE", Bytes: ts.Certificate().Raw}
	if err := os.WriteFile(path, pem.EncodeToMemory(block), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
