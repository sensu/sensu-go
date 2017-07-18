package middlewares

import (
	"context"
	"net/http"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	sensujwt "github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestHandler struct {
	reqCtx context.Context
}

type TestResponseWriter struct {
	header http.Header
}

func (w TestResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}
	return w.header
}

func (w TestResponseWriter) Write(data []byte) (int, error) {
	return 0, nil
}

func (w TestResponseWriter) WriteHeader(status int) {}

func (h *TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.reqCtx = r.Context()
}

func TestAuthorization(t *testing.T) {
	assert := assert.New(t)

	user := &types.User{
		Username: "sensu",
		Password: "passw0rd",
		Roles:    []string{"canhazreadaccess"},
		Disabled: false,
	}

	claims := sensujwt.Claims{
		StandardClaims: jwt.StandardClaims{
			Subject: user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)

	rule := types.Rule{
		Type:         "entities",
		Organization: "sensu",
		Permissions:  []string{types.RulePermRead},
	}

	roles := []*types.Role{
		{
			Name:  "canhazreadaccess",
			Rules: []types.Rule{rule},
		},
	}

	// store needs to return a user and roles
	store := &mockstore.MockStore{}
	store.On("GetUser", mock.Anything).Return(user, nil)
	store.On("GetRoles").Return(roles, nil)

	// create a mock http request w/user context
	req, _ := http.NewRequest("GET", "/foo", nil)
	ctx := sensujwt.SetClaimsIntoContext(req.Context(), token)
	w := TestResponseWriter{}

	// handler needs a context with claims already populated
	next := TestHandler{}
	handler := Authorization(&next, store)
	handler.ServeHTTP(w, req.WithContext(ctx))

	want := roles
	got := next.reqCtx.Value(authorization.ContextRoleKey)

	assert.Equal(want, got)
}
