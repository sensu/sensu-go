package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	etcdstorev2 "github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sirupsen/logrus"
)

type comparable interface {
	Equal(interface{}) bool
}

func testWithEtcdStores(t *testing.T, f func(*etcdstore.Store, *etcdstorev2.Store)) {
	logrus.SetLevel(logrus.ErrorLevel)

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()

	s1 := etcdstore.NewStore(client)
	s2 := etcdstorev2.NewStore(client)

	if err := seeds.SeedInitialDataWithContext(context.Background(), s1); err != nil {
		t.Fatalf("failed to seed initial etcd data: %v", err)
	}

	f(s1, s2)
}

func patchRequest(target, namespace, id, body string) *http.Request {
	r := httptest.NewRequest("PATCH", target, strings.NewReader(body))

	// some of our code reads the namespace from the request context and other
	// code reads it from mux.Vars(), so we must set both.
	r = r.WithContext(store.NamespaceContext(r.Context(), namespace))
	vars := map[string]string{
		"namespace": namespace,
		"id":        id,
	}
	return mux.SetURLVars(r, vars)
}

func TestHandlers_PatchResource(t *testing.T) {
	type fields struct {
		Resource   corev2.Resource
		V3Resource corev3.Resource
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		storeInit func(*testing.T, *etcdstore.Store, *etcdstorev2.Store)
		want      interface{}
		wantErr   bool
	}{
		{
			name: "succeeds & ignores non-existent field for a V2 resource",
			fields: fields{
				Resource: &corev2.CheckConfig{},
			},
			args: args{
				r: patchRequest("/", "default", "testcheck", `{"invalid": ["windows"]}`),
			},
			storeInit: func(t *testing.T, s1 *etcdstore.Store, s2 *etcdstorev2.Store) {
				ctx := store.NamespaceContext(context.Background(), "default")
				check := corev2.FixtureCheckConfig("testcheck")
				if err := s1.UpdateCheckConfig(ctx, check); err != nil {
					t.Fatal(err)
				}
			},
			want: func() interface{} {
				return corev2.FixtureCheckConfig("testcheck")
			}(),
		},
		{
			name: "errors when field has invalid type for a V2 resource",
			fields: fields{
				Resource: &corev2.CheckConfig{},
			},
			args: args{
				r: patchRequest("/", "default", "testcheck", `{"subscriptions": 3}`),
			},
			storeInit: func(t *testing.T, s1 *etcdstore.Store, s2 *etcdstorev2.Store) {
				ctx := store.NamespaceContext(context.Background(), "default")
				check := corev2.FixtureCheckConfig("testcheck")
				if err := s1.UpdateCheckConfig(ctx, check); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: true,
		},
		{
			name: "succeeds when body has valid field for a V2 resource",
			fields: fields{
				Resource: &corev2.CheckConfig{},
			},
			args: args{
				r: patchRequest("/", "default", "testcheck", `{"subscriptions": ["windows"]}`),
			},
			storeInit: func(t *testing.T, s1 *etcdstore.Store, s2 *etcdstorev2.Store) {
				ctx := store.NamespaceContext(context.Background(), "default")
				check := corev2.FixtureCheckConfig("testcheck")
				if err := s1.UpdateCheckConfig(ctx, check); err != nil {
					t.Fatal(err)
				}
			},
			want: func() interface{} {
				check := corev2.FixtureCheckConfig("testcheck")
				check.Subscriptions = []string{"windows"}
				return check
			}(),
		},
		{
			name: "succeeds & ignores non-existent field for a V3 resource",
			fields: fields{
				V3Resource: &corev3.EntityConfig{},
			},
			args: args{
				r: patchRequest("/", "default", "testentity", `{"invalid":["windows"]}`),
			},
			storeInit: func(t *testing.T, s1 *etcdstore.Store, s2 *etcdstorev2.Store) {
				ctx := store.NamespaceContext(context.Background(), "default")
				entity := corev3.FixtureEntityConfig("testentity")
				req := storev2.NewResourceRequestFromResource(ctx, entity)
				wrapper, err := storev2.WrapResource(entity)
				if err != nil {
					t.Fatal(err)
				}
				if err := s2.CreateOrUpdate(req, wrapper); err != nil {
					t.Fatal(err)
				}
			},
			want: func() interface{} {
				return corev3.FixtureEntityConfig("testentity")
			}(),
		},
		{
			name: "errors when field has invalid type for a V3 resource",
			fields: fields{
				V3Resource: &corev3.EntityConfig{},
			},
			args: args{
				r: patchRequest("/", "default", "testentity", `{"subscriptions":3}`),
			},
			storeInit: func(t *testing.T, s1 *etcdstore.Store, s2 *etcdstorev2.Store) {
				ctx := store.NamespaceContext(context.Background(), "default")
				entity := corev3.FixtureEntityConfig("testentity")
				req := storev2.NewResourceRequestFromResource(ctx, entity)
				wrapper, err := storev2.WrapResource(entity)
				if err != nil {
					t.Fatal(err)
				}
				if err := s2.CreateOrUpdate(req, wrapper); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: true,
		},
		{
			name: "succeeds when body has valid field for a V3 resource",
			fields: fields{
				V3Resource: &corev3.EntityConfig{},
			},
			args: args{
				r: patchRequest("/", "default", "testentity", `{"subscriptions":["windows"]}`),
			},
			storeInit: func(t *testing.T, s1 *etcdstore.Store, s2 *etcdstorev2.Store) {
				ctx := store.NamespaceContext(context.Background(), "default")
				entity := corev3.FixtureEntityConfig("testentity")
				req := storev2.NewResourceRequestFromResource(ctx, entity)
				wrapper, err := storev2.WrapResource(entity)
				if err != nil {
					t.Fatal(err)
				}
				if err := s2.CreateOrUpdate(req, wrapper); err != nil {
					t.Fatal(err)
				}
			},
			want: func() interface{} {
				entity := corev3.FixtureEntityConfig("testentity")
				entity.Subscriptions = []string{"windows", "entity:testentity"}
				return entity
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWithEtcdStores(t, func(s1 *etcdstore.Store, s2 *etcdstorev2.Store) {
				if tt.storeInit != nil {
					tt.storeInit(t, s1, s2)
				}
				h := Handlers{
					Resource:   tt.fields.Resource,
					V3Resource: tt.fields.V3Resource,
					Store:      s1,
					StoreV2:    s2,
				}
				got, err := h.PatchResource(tt.args.r)
				if (err != nil) != tt.wantErr {
					t.Errorf("Handlers.PatchResource() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want != nil {
					wantComparable, ok := tt.want.(comparable)
					if !ok {
						t.Fatal("want cannot be type asserted as comparable")
					}
					if !wantComparable.Equal(got) {
						t.Errorf("Handlers.PatchResource() = %v, want %v", got, tt.want)
					}
				}
			})
		})
	}
}
