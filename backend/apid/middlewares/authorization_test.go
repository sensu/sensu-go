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

	claims := types.Claims{
		StandardClaims: jwt.StandardClaims{
			Subject: user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)

	rule := types.Rule{
		Type:        "entities",
		Namespace:   "sensu",
		Permissions: []string{types.RulePermRead},
	}

	roles := []*types.Role{
		{
			Name:  "canhazreadaccess",
			Rules: []types.Rule{rule},
		},
	}

	// store needs to return a user and roles
	store := &mockstore.MockStore{}
	store.On("GetUser", mock.Anything, mock.Anything).Return(user, nil).Once()
	store.On("GetRoles", mock.Anything).Return(roles, nil).Once()

	// create a mock http request w/user context
	req, _ := http.NewRequest("GET", "/foo", nil)
	ctx := sensujwt.SetClaimsIntoContext(req, token.Claims.(*types.Claims))
	w := TestResponseWriter{}

	// handler needs a context with claims already populated
	next := TestHandler{}
	mware := Authorization{Store: store}
	handler := mware.Then(&next)
	handler.ServeHTTP(w, req.WithContext(ctx))

	want := authorization.Actor{Name: "sensu", Rules: roles[0].Rules}
	got := next.reqCtx.Value(types.AuthorizationActorKey)

	assert.Equal(want, got)
}
