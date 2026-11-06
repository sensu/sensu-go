package tessend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSendGivesUpOnUnresponsiveEndpoint(t *testing.T) {
	release := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer ts.Close()
	defer close(release)

	old := sendTimeout
	sendTimeout = 200 * time.Millisecond
	defer func() { sendTimeout = old }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	td := &Tessend{ctx: ctx, url: ts.URL}

	done := make(chan string, 1)
	go func() { done <- td.send(&Data{}) }()
	select {
	case got := <-done:
		if got != "" {
			t.Errorf("send() = %q, want empty on timeout", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("send() blocked on an endpoint that accepts the connection but never responds")
	}
}

func TestSendStopsWhenBackendContextIsCancelled(t *testing.T) {
	release := make(chan struct{})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer ts.Close()
	defer close(release)

	ctx, cancel := context.WithCancel(context.Background())
	td := &Tessend{ctx: ctx, url: ts.URL}
	done := make(chan struct{})
	go func() { td.send(&Data{}); close(done) }()
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("send() ignored the cancelled tessend context")
	}
}
