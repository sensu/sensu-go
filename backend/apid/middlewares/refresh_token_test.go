package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestRefreshTokenNoAccessToken(t *testing.T) {
	mware := RefreshToken{}
	server := httptest.NewServer(mware.Register(testHandler()))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodPost, server.URL, nil)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestRefreshTokenInvalidAccessToken(t *testing.T) {
	mware := RefreshToken{}
	server := httptest.NewServer(mware.Register(testHandler()))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodPost, server.URL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", "foobar"))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TTestRefreshTokenNoRefreshToken(t *testing.T) {
	mware := RefreshToken{}
	server := httptest.NewServer(mware.Register(testHandler()))
	defer server.Close()

	_, tokenString, _ := jwt.AccessToken("foo")

	req, _ := http.NewRequest(http.MethodPost, server.URL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	req, _ = http.NewRequest(http.MethodPost, server.URL, bytes.NewBuffer([]byte("foo")))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	res, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestRefreshTokenInvalidRefreshToken(t *testing.T) {
	mware := RefreshToken{}
	server := httptest.NewServer(mware.Register(testHandler()))
	defer server.Close()

	_, tokenString, _ := jwt.AccessToken("foo")
	refreshTokenString := "foobar"
	body := &types.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, server.URL, bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestRefreshTokenMismatchingSub(t *testing.T) {
	mware := RefreshToken{}
	server := httptest.NewServer(mware.Register(testHandler()))
	defer server.Close()

	_, tokenString, _ := jwt.AccessToken("foo")
	_, refreshTokenString, _ := jwt.RefreshToken("bar")
	body := &types.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, server.URL, bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestRefreshTokenSuccess(t *testing.T) {
	mware := RefreshToken{}
	server := httptest.NewServer(mware.Register(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Make sure the context has been injected with the tokens info
			assert.NotNil(t, r.Context().Value(types.AccessTokenClaims))
			assert.NotNil(t, r.Context().Value(types.RefreshTokenClaims))
			assert.NotNil(t, r.Context().Value(types.RefreshTokenString))

			return
		}),
	))
	defer server.Close()

	_, tokenString, _ := jwt.AccessToken("foo")
	_, refreshTokenString, _ := jwt.RefreshToken("foo")
	body := &types.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, server.URL, bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}
