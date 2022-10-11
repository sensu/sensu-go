package jwt

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/echlebek/crock"
	time "github.com/echlebek/timeproxy"
	jwt "github.com/golang-jwt/jwt/v4"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
)

var testTime = crock.NewTime(time.Now())

func init() {
	time.TimeProxy = testTime
	jwt.TimeFunc = testTime.Now
}

func TestAccessToken(t *testing.T) {
	secret = []byte("foobar")
	claims := &corev2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}

	_, tokenString, err := AccessToken(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, token)

	tokenClaims, _ := token.Claims.(*corev2.Claims)
	assert.Equal(t, claims.Subject, tokenClaims.Subject)
	assert.NotEmpty(t, tokenClaims.Id)
	assert.NotZero(t, tokenClaims.ExpiresAt)
}

func TestClaimsContext(t *testing.T) {
	claims := &corev2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}

	r, _ := http.NewRequest("GET", "/foo", nil)
	ctx := SetClaimsIntoContext(r, claims)

	tokenClaims := GetClaimsFromContext(ctx)
	assert.Equal(t, claims.Subject, tokenClaims.Subject)
}

func TestGetClaims(t *testing.T) {
	claims := &corev2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}
	token, _, _ := AccessToken(claims)

	_, err := GetClaims(token)
	assert.NoError(t, err)
}

func TestExtractBearerToken(t *testing.T) {
	claims := &corev2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}

	// No bearer token
	r, _ := http.NewRequest("GET", "/foo", nil)
	token := ExtractBearerToken(r)

	assert.Empty(t, token)

	// Valid bearer token
	r, _ = http.NewRequest("GET", "/foo", nil)
	_, tokenString, _ := AccessToken(claims)
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))
	token = ExtractBearerToken(r)

	assert.NotEmpty(t, token)
}

func TestInitSecret(t *testing.T) {
	secret = nil
	store := &mockstore.MockStore{}
	store.On("GetJWTSecret").Return("foo", nil)

	err := InitSecret(store)
	assert.NoError(t, err)
	assert.NotEqual(t, nil, secret)
}

func TestInitSecretEtcdError(t *testing.T) {
	secret = nil
	store := &mockstore.MockStore{}
	store.On("GetJWTSecret").Return("", fmt.Errorf(""))
	store.On("CreateJWTSecret").Return(fmt.Errorf(""))

	err := InitSecret(store)
	assert.Error(t, err)
	assert.Equal(t, []byte(nil), secret)
}

func TestRefreshToken(t *testing.T) {
	secret = []byte("foobar")
	claims := &corev2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}
	_, tokenString, err := RefreshToken(claims)
	assert.NoError(t, err)

	token, err := ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, token)

	tokenClaims, _ := token.Claims.(*corev2.Claims)
	assert.Equal(t, claims.Subject, tokenClaims.Subject)
	assert.NotEmpty(t, tokenClaims.Id)
}

func TestValidateTokenError(t *testing.T) {
	// Create an expired token
	claims := &corev2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}
	_, tokenString, _ := AccessToken(claims)

	// Assert that the token is currently valid
	_, err := ValidateToken(tokenString)
	assert.NoError(t, err)

	// The token should expire after the expiration time
	testTime.Set(time.Now().Add(defaultExpiration + time.Hour))
	_, err = ValidateToken(tokenString)
	assert.Error(t, err)
}

// ValidateExpiredToken should not return an error if we provide an expired
// token but otherwise valid
func TestValidateExpiredToken(t *testing.T) {
	// Create an expired token
	claims := &corev2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}
	_, tokenString, _ := AccessToken(claims)

	// Wait for the token to expire
	testTime.Set(time.Now().Add(defaultExpiration + time.Second))
	_, err := ValidateExpiredToken(tokenString)
	assert.NoError(t, err, "An expired token should not be considered as invalid")
}

func TestValidateExpiredTokenActive(t *testing.T) {
	claims := &corev2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}

	_, tokenString, err := AccessToken(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	token, err := ValidateToken(tokenString)
	assert.NoError(t, err, "A valid token should not be considered as invalid")
	assert.NotNil(t, token)
}

func TestValidateExpiredTokenInvalid(t *testing.T) {
	// Create an expired token
	secret = []byte("foobar")
	claims := &corev2.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}
	_, tokenString, _ := AccessToken(claims)

	// The token will expire
	testTime.Set(time.Now().Add(defaultExpiration + time.Second))

	// Modify the secret so it's no longer valid
	secret = []byte("qux")

	_, err := ValidateExpiredToken(tokenString)
	assert.Error(t, err, "An invalid token should not be valid even if it's expired")
}

func TestLoadKeyPair(t *testing.T) {
	tests := []struct {
		name              string
		privatePath       string
		publicPath        string
		wantSigningMethod jwt.SigningMethod
		wantErr           bool
	}{
		{
			name:              "not found private key",
			privatePath:       "testdata/notfound.pem",
			publicPath:        "testdata/ecdsa-p521-public.pem",
			wantSigningMethod: jwt.SigningMethodHS256,
			wantErr:           true,
		},
		{
			name:              "not found public key",
			privatePath:       "testdata/ecdsa-p521-private.pem",
			publicPath:        "testdata/notfound.pem",
			wantSigningMethod: jwt.SigningMethodHS256,
			wantErr:           true,
		},
		{
			name:              "invalid public key",
			privatePath:       "testdata/rsa.pem",
			publicPath:        "testdata/ecdsa-p521-public.pem",
			wantSigningMethod: jwt.SigningMethodHS256,
			wantErr:           true,
		},
		{
			name:              "invalid private key",
			privatePath:       "testdata/ecdsa-p521-private.pem",
			publicPath:        "testdata/rsa.pub",
			wantSigningMethod: jwt.SigningMethodHS256,
			wantErr:           true,
		},
		{
			name:              "valid key pair",
			privatePath:       "testdata/ecdsa-p521-private.pem",
			publicPath:        "testdata/ecdsa-p521-public.pem",
			wantSigningMethod: jwt.SigningMethodES512,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadKeyPair(tt.privatePath, tt.publicPath); (err != nil) != tt.wantErr {
				t.Errorf("LoadKeyPair() error = %v, wantErr %v", err, tt.wantErr)
			}
			if signingMethod != tt.wantSigningMethod {
				t.Fatalf("bad signing method: got %v, want %v", signingMethod, tt.wantSigningMethod)
			}
		})
		// Reset the signing method
		signingMethod = jwt.SigningMethodHS256
	}
}

func TestParseToken(t *testing.T) {
	privateKey = nil
	publicKey = nil
	secret = nil

	type initFunc = func()
	tests := []struct {
		name     string
		token    string
		initFunc initFunc
		// want     *jwt.Token
		wantErr bool
	}{
		{
			name:    "invalid token",
			wantErr: true,
		},
		{
			name:  "symmetric token",
			token: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJmb28iOiJiYXIifQ.ihTPICgltpjcG13pUr_X6KCT8bPPoYZ6Wkm6zk-ramw",
			initFunc: func() {
				secret = []byte("P@ssw0rd!")
			},
		},
		{
			name:  "invalid symmetric token",
			token: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJmb28iOiJiYXIifQ.ihTPICgltpjcG13pUr_X6KCT8bPPoYZ6Wkm6zk-ramw",
			initFunc: func() {
				secret = []byte("invalid_secret")
			},
			wantErr: true,
		},
		{
			name:  "invalid symmetric algorithm",
			token: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzUxMiJ9.eyJmb28iOiJiYXIifQ.QYsHpDgSnAUsx1ZESOWDmHasix9ZpdiL8fYhHVvKryNq-1R6xnlVDqc43NZjG8X682dWt40QFjLpMa1IWE0nwQ",
			initFunc: func() {
				secret = []byte("P@ssw0rd!")
			},
			wantErr: true,
		},
		{
			name:    "no asymmetric public key",
			token:   "eyJ0eXAiOiJKV1QiLCJhbGciOiJFUzI1NiJ9.eyJmb28iOiJiYXIifQ.Q50GauuUNxJcUiW1Ss3sxCezmVsHYuRHcaQNrZI6iHGpyMTkhx_YRUHvm-s7GdInWyjdBo5OrjuJW_NTDr6xHw",
			wantErr: true,
		},
		{
			name:  "invalid asymmetric algorithm",
			token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.POstGetfAytaZS82wHcjoTyoqhMyxXiWdR7Nn7A29DNSl0EiXLdwJ6xC6AfgZWF1bOsS_TuYI3OG85AmiExREkrS6tDfTQ2B3WXlrr-wp5AokiRbz3_oB4OxG-W9KcEEbDRcZc0nH3L7LzYptiy1PtAylQGxHTWZXtGz4ht0bAecBgmpdgXMguEIcoqPJ1n3pIWk_dUZegpqx0Lka21H6XxUTxiy8OcaarA8zdnPUnV6AmNP3ecFawIFYdvJB_cm-GvpCSbr8G8y_Mllj8f4x9nBH8pQux89_6gUY618iYv7tuPWBFfEbLxtF2pZS6YC1aSfLQxeNe8djT9YjpvRZA",
			initFunc: func() {
				publicKey, _ = jwt.ParseECPublicKeyFromPEM([]byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAET/mexGk7Q1GuhI0OnulJZOWdrql7
xCMF9dzNuX/Brbdf5i9c90eqTD4LBUDAfkr5MsXHRs2MIQsS5waVzy6Q9A==
-----END PUBLIC KEY-----`))
			},
			wantErr: true,
		},
		{
			name:  "asymmetric token",
			token: "eyJ0eXAiOiJKV1QiLCJhbGciOiJFUzI1NiJ9.eyJmb28iOiJiYXIifQ.Q50GauuUNxJcUiW1Ss3sxCezmVsHYuRHcaQNrZI6iHGpyMTkhx_YRUHvm-s7GdInWyjdBo5OrjuJW_NTDr6xHw",
			initFunc: func() {
				publicKey, _ = jwt.ParseECPublicKeyFromPEM([]byte(`-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAET/mexGk7Q1GuhI0OnulJZOWdrql7
xCMF9dzNuX/Brbdf5i9c90eqTD4LBUDAfkr5MsXHRs2MIQsS5waVzy6Q9A==
-----END PUBLIC KEY-----`))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initFunc != nil {
				tt.initFunc()
			}

			_, err := parseToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
