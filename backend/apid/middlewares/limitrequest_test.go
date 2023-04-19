package middlewares

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareLimits(t *testing.T) {
	assert := assert.New(t)

	goodCheck := &v2.CheckConfig{
		ObjectMeta: v2.ObjectMeta{
			Name:      "goodcheck",
			Namespace: "default",
		},
		Command:       "true",
		Interval:      30,
		Publish:       true,
		Subscriptions: []string{"system"},
	}

	maxCheck := make([]byte, 600000)
	rand.Read(maxCheck)
	badCheck := &v2.CheckConfig{
		ObjectMeta: v2.ObjectMeta{
			Name:      "badcheck",
			Namespace: "default",
		},
		Command:       string(maxCheck),
		Interval:      30,
		Publish:       true,
		Subscriptions: []string{"system"},
	}

	tests := []struct {
		description  string
		url          string
		body         *v2.CheckConfig
		expectedCode int
		limit        int64
	}{
		{
			description:  "Request within threshold",
			url:          "/checks",
			body:         goodCheck,
			expectedCode: http.StatusOK,
			limit:        MaxBytesLimit,
		}, {
			description:  "Request over threshold",
			url:          "/checks",
			body:         badCheck,
			expectedCode: http.StatusInternalServerError,
			limit:        MaxBytesLimit,
		}, {
			description:  "Configurable limit within threshold",
			url:          "/checks",
			body:         goodCheck,
			expectedCode: http.StatusOK,
			limit:        1024000,
		}, {
			description:  "Configurable limit over threshold",
			url:          "/checks",
			body:         goodCheck,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		mware := LimitRequest{Limit: tc.limit}
		server := httptest.NewServer(mware.Then(testHandler()))
		defer server.Close()
		payload, _ := json.Marshal(tc.body)
		req, _ := http.NewRequest(http.MethodPost, server.URL+tc.url, bytes.NewBuffer(payload))
		res, err := http.DefaultClient.Do(req)
		assert.NoError(err)
		assert.Equal(tc.expectedCode, res.StatusCode, tc.description)
	}
}
