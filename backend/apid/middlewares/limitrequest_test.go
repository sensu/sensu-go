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

func TestMiddleWareLimitRequest(t *testing.T) {
	limit := LimitRequest{}
	server := httptest.NewServer(limit.Then(testHandler()))
	defer server.Close()

	check := &types.CheckConfig{
		Command:       "true",
		Environment:   "default",
		Interval:      30,
		Name:          "checktest",
		Organization:  "default",
		Publish:       true,
		Subscriptions: []string{"system"},
	}

	payload, _ := json.Marshal(check)
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/checks", bytes.NewBuffer(payload))
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestMiddleWareInvalidLimitRequest(t *testing.T) {
	limit := LimitRequest{}
	server := httptest.NewServer(limit.Then(testHandler()))
	defer server.Close()

	maxCheck := make([]byte, 600000)
	rand.Read(maxCheck)
	check := &types.CheckConfig{
		Command:       string(maxCheck),
		Environment:   "default",
		Interval:      30,
		Name:          "checktest",
		Organization:  "default",
		Publish:       true,
		Subscriptions: []string{"system"},
	}

	payload, _ := json.Marshal(check)
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/checks", bytes.NewBuffer(payload))
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
