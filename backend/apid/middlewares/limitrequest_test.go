package middlewares

import (
	"fmt"
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
	fmt.Println(len(string(maxCheck)))
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
	fmt.Println(checkBody.Len())
	fmt.Println(server.URL)
	fmt.Println(req.URL)
	res, err := client.Do(req)
	fmt.Println(res)
	assert.Error(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}
