package routers

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func newRequest(t *testing.T, method, endpoint string, body io.Reader) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		t.Fatal(err)
	}

	return req.WithContext(context.Background())
}
