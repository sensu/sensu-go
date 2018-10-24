package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	jwt "github.com/dgrijalva/jwt-go"
	sensujwt "github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/types"
)

type FakeAuthorizer struct {
	authorized bool
	err        error
}

func (a *FakeAuthorizer) Authorize(attrs *authorization.Attributes) (bool, error) {
	if a.err != nil {
		return false, a.err
	}
	return a.authorized, nil
}

func TestAuthorization(t *testing.T) {
	cases := []struct {
		name            string
		claims          *types.Claims
		attrs           *authorization.Attributes
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
			attrs:           &authorization.Attributes{},
			authorizedError: errors.New("error"),
			expected:        http.StatusInternalServerError,
		},
		{
			name:         "unauthorized",
			claims:       &types.Claims{},
			attrs:        &authorization.Attributes{},
			isAuthorized: false,
			expected:     http.StatusForbidden,
		},
		{
			name:         "unauthorized",
			claims:       &types.Claims{},
			attrs:        &authorization.Attributes{},
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
			ctx = authorization.SetAttributes(ctx, tt.attrs)

			// serve our request with our authorization middleware
			handler := authzMiddleware.Then(&testHandler)
			handler.ServeHTTP(w, r.WithContext(ctx))

			assert.Equal(t, tt.expected, w.Code)
		})
	}

}
