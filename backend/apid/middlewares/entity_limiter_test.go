package middlewares

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/limiter"
	"github.com/stretchr/testify/assert"
)

func TestMiddlewareEntityLimit(t *testing.T) {
	assert := assert.New(t)
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	tests := []struct {
		description string
		entities    int
		header      bool
	}{
		{
			description: "Entities over limit",
			entities:    1500,
			header:      true,
		}, {
			description: "Entities under limit",
			entities:    500,
			header:      false,
		}, {
			description: "Entities at limit",
			entities:    1000,
			header:      false,
		},
	}

	mware := EntityLimiter{
		Limiter: limiter.NewEntityLimiter(context.Background(), client),
	}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	for _, tc := range tests {
		for i := 0; i < 12; i++ {
			mware.Limiter.AddCount(tc.entities)
		}
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/health", bytes.NewBuffer([]byte{}))
		res, err := http.DefaultClient.Do(req)
		assert.NoError(err)
		assert.Equal(http.StatusOK, res.StatusCode, tc.description)
		if tc.header {
			assert.Contains(res.Header.Get(HeaderWarning), "You have exceeded the entity limit", tc.description)
		} else {
			assert.Empty(res.Header.Get(HeaderWarning), tc.description)
		}
	}
}
