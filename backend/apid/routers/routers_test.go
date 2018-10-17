package routers

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/types"
)

func newRequest(t *testing.T, method, endpoint string, body io.Reader) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	rules := []types.Rule{
		types.Rule{
			Type:        types.RuleTypeAll,
			Namespace:   "*",
			Permissions: types.RuleAllPerms,
		},
	}
	actor := authorization.Actor{
		Name:  "admin",
		Rules: rules,
	}
	ctx = context.WithValue(ctx, types.AuthorizationActorKey, actor)
	return req.WithContext(ctx)
}
