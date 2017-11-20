package middlewares

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddleWareLimitRequest(t *testing.T) {
	limit := LimitRequest{}
	server := httptest.NewServer(limit.Then(testHandler()))
	defer server.Close()

	client := &http.Client{}
	checkBody := strings.NewReader(`{
		"Command": 				"true",
		"Environment": 		"default",
		"Interval": 			30,
		"Name":         	"checktest",
		"Organization": 	"default",
		"Publish":      	true,
		"Subscriptions":	[]string{"system"}
	}`)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/checks", checkBody)
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestMiddleWareInvalidLimitRequest(t *testing.T) {
	limit := LimitRequest{}
	server := httptest.NewServer(limit.Then(testHandler()))
	defer server.Close()

	client := &http.Client{}
	maxCheck := make([]byte, 600000)
	rand.Read(maxCheck)
	checkBody := strings.NewReader(`{
		"Command": 				` + string(maxCheck) + `,
		"Environment": 		"default",
		"Interval": 			30,
		"Name":         	"checktest",
		"Organization": 	"default",
		"Publish":      	true,
		"Subscriptions":	[]string{"system"}
	}`)

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/checks", checkBody)
	res, err := client.Do(req)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
