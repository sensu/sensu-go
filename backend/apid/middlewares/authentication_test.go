package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMiddlewareNoCredentials(t *testing.T) {
	mware := Authentication{}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	// No credentials passed
	res, err := http.Get(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestMiddlewareJWT(t *testing.T) {
	mware := Authentication{}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	// Valid JWT
	claims := corev2.FixtureClaims("foo", nil)
	_, tokenString, _ := jwt.AccessToken(claims)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)

	// Add the bearer token in the Authorization header
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestMiddlewareInvalidJWT(t *testing.T) {
	mware := Authentication{}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	// Valid JWT
	tokenString := "foobar"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)

	// Add the bearer token in the Authorization header
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestMiddlewareIgnoreUnauthorized(t *testing.T) {
	mware := Authentication{IgnoreUnauthorized: true}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	// No credentials passed
	res, err := http.Get(server.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestMiddlewareValidAPIKey(t *testing.T) {
	store := &mockstore.V2MockStore{}
	cs := new(mockstore.ConfigStore)
	store.On("GetConfigStore").Return(cs)
	mware := Authentication{
		Store: store,
	}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	key := corev2.FixtureAPIKey("174373d0-4aff-41d8-aa5f-084dfcad7dc7", "admin")
	keyReq := storev2.NewResourceRequestFromResource(key)
	user := &corev2.User{Username: "admin"}
	userReq := storev2.NewResourceRequestFromResource(user)
	cs.On("Get", mock.Anything, keyReq).Return(mockstore.Wrapper[*corev2.APIKey]{Value: key}, nil)
	cs.On("Get", mock.Anything, userReq).Return(mockstore.Wrapper[*corev2.User]{Value: user}, nil)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Key %s", key.Name))
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestMiddlewareInvalidAPIKey(t *testing.T) {
	store := &mockstore.V2MockStore{}
	cs := new(mockstore.ConfigStore)
	store.On("GetConfigStore").Return(cs)
	mware := Authentication{
		Store: store,
	}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	key := corev2.FixtureAPIKey("174373d0-4aff-41d8-aa5f-084dfcad7dc7", "admin")
	cs.On("Get", mock.Anything, mock.Anything).Return(nil, errors.New("err"))

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Key %s", key.Name))
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestMiddlewareInvalidUserAPIKey(t *testing.T) {
	store := &mockstore.V2MockStore{}
	cs := new(mockstore.ConfigStore)
	store.On("GetConfigStore").Return(cs)
	mware := Authentication{
		Store: store,
	}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	key := corev2.FixtureAPIKey("174373d0-4aff-41d8-aa5f-084dfcad7dc7", "admin")
	keyReq := storev2.NewResourceRequestFromResource(key)
	user := &corev2.User{Username: "admin"}
	userReq := storev2.NewResourceRequestFromResource(user)
	cs.On("Get", mock.Anything, keyReq).Return(mockstore.Wrapper[*corev2.APIKey]{Value: key}, nil)
	cs.On("Get", mock.Anything, userReq).Return(nil, errors.New("err"))

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Key %s", key.Name))
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestMiddlewareNoUserAPIKey(t *testing.T) {
	st := &mockstore.V2MockStore{}
	cs := new(mockstore.ConfigStore)
	st.On("GetConfigStore").Return(cs)
	mware := Authentication{
		Store: st,
	}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	key := corev2.FixtureAPIKey("174373d0-4aff-41d8-aa5f-084dfcad7dc7", "admin")
	keyReq := storev2.NewResourceRequestFromResource(key)
	user := &corev2.User{Username: "admin"}
	userReq := storev2.NewResourceRequestFromResource(user)
	cs.On("Get", mock.Anything, keyReq).Return(nil, &store.ErrNotFound{})
	cs.On("Get", mock.Anything, userReq).Return(mockstore.Wrapper[*corev2.User]{Value: user}, nil)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Key %s", key.Name))
	res, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}
