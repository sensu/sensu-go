package api

import (
	"context"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
)

func defaultContext() context.Context {
	return store.NamespaceContext(context.Background(), "default")
}

func TestListAssets(t *testing.T) {
	tests := []struct {
		Name   string
		Ctx    func() context.Context
		Store  func() store.Store
		Exp    []*corev2.Asset
		ExpErr bool
	}{
		{
			Name: "no auth",
			Ctx:  defaultContext,
			Store: func() store.Store {
				store := new(mockstore.MockStore)
				return store
			},
			ExpErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := test.Ctx()
			store := test.Store()
			auth := &rbac.Authorizer{Store: store}
			client := NewAssetClient(store, auth)
			assets, err := client.ListAssets(ctx)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected non-nil error")
			}
			if got, want := assets, test.Exp; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad assets: got %v, want %v", got, want)
			}
		})
	}
}
