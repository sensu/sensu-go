package middlewares

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareLimits(t *testing.T) {
	assert := assert.New(t)

	goodCheck := &types.CheckConfig{
		ObjectMeta: types.ObjectMeta{
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
	badCheck := &types.CheckConfig{
		ObjectMeta: types.ObjectMeta{
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
		body         *types.CheckConfig
		expectedCode int
	}{
		{
			description:  "Request within threshold",
			url:          "/checks",
			body:         goodCheck,
			expectedCode: http.StatusOK,
		}, {
			description:  "Request over threshold",
			url:          "/checks",
			body:         badCheck,
			expectedCode: http.StatusInternalServerError,
		},
	}

	mware := LimitRequest{}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	for _, tc := range tests {
		payload, _ := json.Marshal(tc.body)
		req, _ := http.NewRequest(http.MethodPost, server.URL+tc.url, bytes.NewBuffer(payload))
		res, err := http.DefaultClient.Do(req)
		assert.NoError(err)
		assert.Equal(tc.expectedCode, res.StatusCode, tc.description)
	}
}
