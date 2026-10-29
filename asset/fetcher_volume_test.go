package asset

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// The volume tests check that a high number of asset fetches does not grow the
// process's file descriptor table or the server-side connection count. They are
// tuned with environment variables, so the same file runs quickly in CI and at
// full volume on a dedicated host:
//
//	ASSET_VOLUME_WORKERS     concurrent fetchers per scenario (default 20)
//	ASSET_VOLUME_PER_WORKER  requests per fetcher per scenario (default 250)
//	ASSET_VOLUME_FD_BOUND    allowed growth in open file descriptors (default 50)

const (
	expectMixed = iota
	expectOK
	expectFail
)

// maxIdleServerConns is the number of connections a client may keep open after
// a burst. The default transport keeps DefaultMaxIdleConnsPerHost (2) idle
// connections per host, so anything above a small slack means connections are
// leaking rather than being pooled or closed.
const maxIdleServerConns = 5

func volumeInt(name string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(name)); err == nil && v > 0 {
		return v
	}
	return def
}

// openFDCount returns the number of open file descriptors in this process. It
// skips the test where the platform does not expose them.
func openFDCount(t *testing.T) int {
	t.Helper()
	for _, dir := range []string{"/proc/self/fd", "/dev/fd"} {
		if entries, err := os.ReadDir(dir); err == nil {
			return len(entries)
		}
	}
	t.Skip("file descriptor counting needs /proc/self/fd or /dev/fd")
	return 0
}

type volumeScenario struct {
	name    string
	useTLS  bool
	useCA   bool
	timeout time.Duration
	expect  int
	handler http.HandlerFunc
}

func volumeScenarios() []volumeScenario {
	var mixed atomic.Int64

	return []volumeScenario{
		{
			name:   "http_404_no_ca",
			expect: expectFail,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("not found"))
			},
		},
		{
			name:   "http_200_no_ca",
			expect: expectOK,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Write([]byte("asset"))
			},
		},
		{
			name:   "tls_404_with_ca",
			useTLS: true,
			useCA:  true,
			expect: expectFail,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("not found"))
			},
		},
		{
			name:   "tls_200_with_ca",
			useTLS: true,
			useCA:  true,
			expect: expectOK,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Write([]byte("asset"))
			},
		},
		{
			// The server asks for the connection to close after every response,
			// so the client gets no keep-alive reuse at all.
			name:   "server_connection_close",
			useTLS: true,
			useCA:  true,
			expect: expectOK,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Connection", "close")
				w.Write([]byte("asset"))
			},
		},
		{
			// The server promises more body than it sends and then aborts. The
			// client gets a 200 and must release the connection when the body is
			// closed early.
			name:   "mid_body_abort",
			useTLS: true,
			useCA:  true,
			expect: expectOK,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Length", "1000")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("partial"))
				// Flush so the 200 and partial body reach the client before the
				// connection is aborted. Without it the client sees EOF before
				// any response and never reaches the body.
				w.(http.Flusher).Flush()
				panic(http.ErrAbortHandler)
			},
		},
		{
			// Alternating success and 404 on one keep-alive pool, which is the
			// shape of a real agent running healthy and missing-asset checks.
			name:   "mixed_success_and_404",
			useTLS: true,
			useCA:  true,
			expect: expectMixed,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				if mixed.Add(1)%2 == 0 {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.Write([]byte("asset"))
			},
		},
		{
			// Each request gives up before the slow server responds. Cancelled
			// requests must release their connections too.
			name:    "context_cancel_slow_server",
			useTLS:  true,
			useCA:   true,
			timeout: 10 * time.Millisecond,
			expect:  expectFail,
			handler: func(w http.ResponseWriter, _ *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.Write([]byte("asset"))
			},
		},
	}
}

// TestHTTPGetHighVolume drives each scenario with concurrent fetchers and
// checks three things: the outcome of every fetch, the growth in open file
// descriptors, and the number of server-side connections still open afterwards.
func TestHTTPGetHighVolume(t *testing.T) {
	workers := volumeInt("ASSET_VOLUME_WORKERS", 20)
	perWorker := volumeInt("ASSET_VOLUME_PER_WORKER", 250)
	fdBound := volumeInt("ASSET_VOLUME_FD_BOUND", 50)
	total := int64(workers * perWorker)

	for _, sc := range volumeScenarios() {
		sc := sc
		t.Run(sc.name, func(t *testing.T) {
			t.Cleanup(resetClientCache)
			tracker := &openConnTracker{}
			ts := httptest.NewUnstartedServer(sc.handler)
			ts.Config.ConnState = tracker.hook
			if sc.useTLS {
				ts.StartTLS()
			} else {
				ts.Start()
			}
			defer ts.Close()

			caFile := ""
			if sc.useCA {
				caFile = writeServerCA(t, ts)
			}

			// Warm up so the baseline already includes the cached client and its
			// first pooled connection.
			if rc, err := httpGet(context.Background(), ts.URL, caFile, nil); err == nil {
				rc.Close()
			}
			time.Sleep(50 * time.Millisecond)

			timeout := sc.timeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}

			before := openFDCount(t)
			var okCount, failCount atomic.Int64
			var wg sync.WaitGroup
			start := time.Now()
			for w := 0; w < workers; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < perWorker; i++ {
						ctx, cancel := context.WithTimeout(context.Background(), timeout)
						rc, err := httpGet(ctx, ts.URL, caFile, nil)
						if err == nil {
							rc.Close()
							okCount.Add(1)
						} else {
							failCount.Add(1)
						}
						cancel()
					}
				}()
			}
			wg.Wait()
			elapsed := time.Since(start)

			waitForOpenConns(tracker, maxIdleServerConns, 2*time.Second)
			after := openFDCount(t)
			delta := after - before
			serverOpen := tracker.count()

			t.Logf("scenario=%s requests=%d ok=%d failed=%d elapsed=%s fds before=%d after=%d delta=%d server_open_conns=%d",
				sc.name, total, okCount.Load(), failCount.Load(), elapsed.Round(time.Millisecond),
				before, after, delta, serverOpen)

			assert.LessOrEqual(t, delta, fdBound, "file descriptor growth after %d requests", total)
			assert.LessOrEqual(t, serverOpen, maxIdleServerConns, "server-side open connections after %d requests", total)

			switch sc.expect {
			case expectOK:
				assert.Equal(t, total, okCount.Load(), "every fetch should succeed")
			case expectFail:
				assert.Equal(t, int64(0), okCount.Load(), "every fetch should fail")
			}
		})
	}
}
