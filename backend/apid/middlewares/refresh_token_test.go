package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/stretchr/testify/assert"
)

func TestRefreshTokenNoAccessToken(t *testing.T) {
	mware := RefreshToken{}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodPost, server.URL, nil)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestRefreshTokenInvalidAccessToken(t *testing.T) {
	mware := RefreshToken{}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	req, _ := http.NewRequest(http.MethodPost, server.URL, nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", "foobar"))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestRefreshTokenNoRefreshToken(t *testing.T) {
	mware := RefreshToken{}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	claims := corev2.FixtureClaims("foo", nil)
	_, tokenString, _ := jwt.AccessToken(claims)

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
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	claims := corev2.FixtureClaims("foo", nil)
	_, tokenString, _ := jwt.AccessToken(claims)
	refreshTokenString := "foobar"
	body := &corev2.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, server.URL, bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestRefreshTokenMismatchingSub(t *testing.T) {
	mware := RefreshToken{}
	server := httptest.NewServer(mware.Then(testHandler()))
	defer server.Close()

	fooClaims := corev2.FixtureClaims("foo", nil)
	barClaims := corev2.FixtureClaims("bar", nil)
	_, tokenString, _ := jwt.AccessToken(fooClaims)
	_, refreshTokenString, _ := jwt.RefreshToken(barClaims)

	body := &corev2.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, server.URL, bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestRefreshTokenSuccess(t *testing.T) {
	mware := RefreshToken{}
	server := httptest.NewServer(mware.Then(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Make sure the context has been injected with the tokens info
			if r.Context().Value(corev2.AccessTokenClaims) == nil {
				http.Error(w, "nil AccessTokenClaims", 500)
				return
			}
			if r.Context().Value(corev2.RefreshTokenClaims) == nil {
				http.Error(w, "nil RefreshTokenClaims", 500)
				return
			}
			if r.Context().Value(corev2.RefreshTokenString) == nil {
				http.Error(w, "nil RefreshTokenString", 500)
			}
		}),
	))
	defer server.Close()

	claims := corev2.FixtureClaims("foo", nil)
	_, tokenString, _ := jwt.AccessToken(claims)
	_, refreshTokenString, _ := jwt.RefreshToken(claims)
	body := &corev2.Tokens{Refresh: refreshTokenString}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, server.URL, bytes.NewBuffer(payload))
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	b, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, "", string(b))
}
