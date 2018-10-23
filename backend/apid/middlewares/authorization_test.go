package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	jwt "github.com/dgrijalva/jwt-go"
	sensujwt "github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/types"
)

type FakeAuthorizer struct {
	authorized bool
	err        error
}

func (a *FakeAuthorizer) Authorize(reqInfo *types.RequestInfo) (bool, error) {
	if a.err != nil {
		return false, a.err
	}
	return a.authorized, nil
}

func TestAuthorization(t *testing.T) {
	cases := []struct {
		name            string
		claims          *types.Claims
		reqInfo         *types.RequestInfo
		authorizedError error
		isAuthorized    bool
		expected        int
	}{
		{
			name:     "missing JWT claims & request info",
			expected: http.StatusInternalServerError,
		},
		{
			name:     "missing request info",
			claims:   &types.Claims{},
			expected: http.StatusInternalServerError,
		},
		{
			name:            "authorizer error",
			claims:          &types.Claims{},
			reqInfo:         &types.RequestInfo{},
			authorizedError: errors.New("error"),
			expected:        http.StatusInternalServerError,
		},
		{
			name:         "unauthorized",
			claims:       &types.Claims{},
			reqInfo:      &types.RequestInfo{},
			isAuthorized: false,
			expected:     http.StatusForbidden,
		},
		{
			name:         "unauthorized",
			claims:       &types.Claims{},
			reqInfo:      &types.RequestInfo{},
			isAuthorized: true,
			expected:     http.StatusOK,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			fakeAuthorizer := &FakeAuthorizer{
				authorized: tt.isAuthorized,
				err:        tt.authorizedError,
			}
			authzMiddleware := Authorization{Authorizer: fakeAuthorizer}
			w := httptest.NewRecorder()
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			// prepare our http request
			r, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal("Couldn't create request: ", err)
			}

			// inject the claims into the request context
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, tt.claims)
			ctx := sensujwt.SetClaimsIntoContext(r, token.Claims.(*types.Claims))

			// inject the request info into the request context
			ctx = context.WithValue(ctx, types.RequestInfoKey, tt.reqInfo)

			// serve our request with our authorization middleware
			handler := authzMiddleware.Then(&testHandler)
			handler.ServeHTTP(w, r.WithContext(ctx))

			assert.Equal(t, tt.expected, w.Code)
		})
	}

}
