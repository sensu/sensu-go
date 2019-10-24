package middlewares

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareEntityLimit(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		description  string
		limitReached bool
	}{
		{
			description:  "Limit has not been reached",
			limitReached: false,
		}, {
			description:  "Limit has been reached",
			limitReached: true,
		},
	}

	for _, tc := range tests {
		mware := EntityLimiter{}
		testMware := testMiddleware{
			limitReached: tc.limitReached,
		}
		var server *httptest.Server
		server = httptest.NewServer(testMware.then(mware.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = fmt.Fprint(w, "Success")
		}))))
		defer server.Close()
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/entities", bytes.NewBuffer([]byte{}))
		res, err := http.DefaultClient.Do(req)
		assert.NoError(err)
		if tc.limitReached {
			assert.Equal(http.StatusPaymentRequired, res.StatusCode, tc.description)
		} else {
			assert.Equal(http.StatusOK, res.StatusCode, tc.description)
		}
	}
}

type testMiddleware struct {
	limitReached bool
}

func (t testMiddleware) then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if t.limitReached {
			r.Header.Set(helpers.HeaderWarning, "We noticed you've reached the blah blah blah")
		}
		next.ServeHTTP(w, r)
	})
}
