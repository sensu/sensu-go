package api

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/storetest"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func defaultTestClient(store storev2.Interface, auth authorization.Authorizer) *GenericClient {
	return &GenericClient{
		Kind:       defaultResource(),
		Store:      store,
		Auth:       auth,
		APIGroup:   "core",
		APIVersion: "v2",
	}
}

func defaultV2TestClient(store storev2.Interface, auth authorization.Authorizer) *GenericClient {
	return &GenericClient{
		Kind:       defaultV3Resource(),
		Store:      store,
		Auth:       auth,
		APIGroup:   "core",
		APIVersion: "v3",
	}
}

func defaultV2ResourceStore() storev2.Interface {
	store := new(storetest.Store)
	wrappedResource, err := storev2.WrapResource(corev3.FixtureEntityConfig("default"))
	if err != nil {
		panic(err)
	}
	store.On("Get", mock.Anything, mock.Anything).Return(wrappedResource, nil)
	store.On("List", mock.Anything, mock.Anything, mock.Anything).Return(wrap.List{wrappedResource.(*wrap.Wrapper)}, nil)
	store.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	store.On("CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	store.On("Delete", mock.Anything, mock.Anything).Return(nil)
	return store
}

func defaultV3Resource() corev3.Resource {
	return corev3.FixtureEntityConfig("default")
}

type mockAuth struct {
	attrs map[authorization.AttributesKey]bool
}

func defaultResource() corev3.Resource {
	return corev2.FixtureAsset("default")
}

func (m *mockAuth) Authorize(ctx context.Context, attrs *authorization.Attributes) (bool, error) {
	if attrs == nil {
		return false, nil
	}
	if corev2.ContextNamespace(ctx) != attrs.Namespace {
		return false, errors.New("namespace does not match")
	}
	b, ok := m.attrs[attrs.Key()]
	if !ok {
		return false, fmt.Errorf("auth missing for %v", attrs.Key())
	}
	return b, nil
}

func badAuth() authorization.Authorizer {
	return &mockAuth{}
}

func contextWithUser(ctx context.Context, username string, groups []string) context.Context {
	return context.WithValue(ctx, corev2.ClaimsKey, corev2.FixtureClaims(username, groups))
}

func defaultResourceStore() storev2.Interface {
	store := &mockstore.V2MockStore{}
	store.On("Get", mock.Anything, mock.Anything).Return(mockstore.Wrapper[*corev2.Asset]{Value: corev2.FixtureAsset("default")}, nil)
	store.On("List", mock.Anything, mock.Anything, mock.Anything).Return(mockstore.WrapList[corev3.Resource]{defaultResource()}, nil)
	store.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	store.On("CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	store.On("Delete", mock.Anything, mock.Anything).Return(nil)
	return store
}

func TestGenericClient(t *testing.T) {
	tests := []struct {
		Name      string
		Client    *GenericClient
		CreateVal corev3.Resource
		CreateErr bool
		UpdateVal corev3.Resource
		UpdateErr bool
		GetName   string
		GetVal    corev3.Resource
		GetErr    bool
		ListVal   []corev3.Resource
		ListPred  *store.SelectionPredicate
		ListErr   bool
		DelName   string
		DelErr    bool
		AuthVerb  RBACVerb
		AuthName  string
		AuthErr   bool
		Ctx       context.Context
	}{
		{
			Name:      "authorizer rejects all",
			Client:    defaultTestClient(defaultResourceStore(), badAuth()),
			CreateVal: defaultResource(),
			CreateErr: true,
			UpdateVal: defaultResource(),
			UpdateErr: true,
			GetName:   "toget",
			GetVal:    defaultResource(),
			GetErr:    true,
			DelName:   "todelete",
			DelErr:    true,
			ListErr:   true,
			AuthVerb:  VerbGet,
			AuthName:  "todelete",
			AuthErr:   true,
			Ctx:       contextWithUser(defaultContext(), "tom", nil),
		},
		{
			Name: "readonly access",
			Client: defaultTestClient(defaultResourceStore(), &mockAuth{
				attrs: map[authorization.AttributesKey]bool{
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v2",
						Namespace:    "default",
						Resource:     "assets",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "get",
					}: true,
					authorization.AttributesKey{
						APIGroup:   "core",
						APIVersion: "v2",
						Namespace:  "default",
						Resource:   "assets",
						UserName:   "tom",
						Verb:       "list",
					}: true,
				},
			}),
			CreateVal: defaultResource(),
			CreateErr: true,
			UpdateVal: defaultResource(),
			UpdateErr: true,
			GetName:   "default",
			GetVal:    defaultResource(),
			GetErr:    false,
			DelName:   "default",
			DelErr:    true,
			ListErr:   false,
			AuthVerb:  VerbDelete,
			AuthName:  "default",
			AuthErr:   true,
			Ctx:       contextWithUser(defaultContext(), "tom", nil),
		},
		{
			Name: "all access",
			Client: defaultTestClient(defaultResourceStore(), &mockAuth{
				attrs: map[authorization.AttributesKey]bool{
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v2",
						Namespace:    "default",
						Resource:     "assets",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "get",
					}: true,
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v2",
						Namespace:    "default",
						Resource:     "assets",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "create",
					}: true,
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v2",
						Namespace:    "default",
						Resource:     "assets",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "delete",
					}: true,
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v2",
						Namespace:    "default",
						Resource:     "assets",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "update",
					}: true,
					authorization.AttributesKey{
						APIGroup:   "core",
						APIVersion: "v2",
						Namespace:  "default",
						Resource:   "assets",
						UserName:   "tom",
						Verb:       "list",
					}: true,
				},
			}),
			CreateVal: defaultResource(),
			CreateErr: false,
			UpdateVal: defaultResource(),
			UpdateErr: false,
			GetName:   "default",
			GetVal:    defaultResource(),
			GetErr:    false,
			DelName:   "default",
			DelErr:    false,
			ListErr:   false,
			AuthVerb:  VerbGet,
			AuthName:  "default",
			AuthErr:   false,
			Ctx:       contextWithUser(defaultContext(), "tom", nil),
		},
	}

	for _, test := range tests {
		t.Run(test.Name+"_create", func(t *testing.T) {
			err := test.Client.Create(test.Ctx, test.CreateVal)
			if err != nil && !test.CreateErr {
				t.Fatal(err)
			}
			if err == nil && test.CreateErr {
				t.Fatal("expected non-nil error")
			}
		})
		t.Run(test.Name+"_update", func(t *testing.T) {
			err := test.Client.Update(test.Ctx, test.UpdateVal)
			if err != nil && !test.UpdateErr {
				t.Fatal(err)
			}
			if err == nil && test.UpdateErr {
				t.Fatal(err)
			}
		})
		t.Run(test.Name+"_get", func(t *testing.T) {
			err := test.Client.Get(test.Ctx, test.GetName, test.GetVal)
			if err != nil && !test.GetErr {
				t.Fatal(err)
			}
			if err == nil && test.GetErr {
				t.Fatal(err)
			}
			if err == nil && test.GetVal.Validate() != nil {
				t.Fatal(test.GetVal.Validate())
			}
		})
		t.Run(test.Name+"_del", func(t *testing.T) {
			err := test.Client.Delete(test.Ctx, test.DelName)
			if err != nil && !test.DelErr {
				t.Fatal(err)
			}
			if err == nil && test.DelErr {
				t.Fatal(err)
			}
		})
		t.Run(test.Name+"_list", func(t *testing.T) {
			err := test.Client.List(test.Ctx, &test.ListVal, test.ListPred)
			if err != nil && !test.ListErr {
				t.Fatal(err)
			}
			if err == nil && test.ListErr {
				t.Fatal(err)
			}
			for _, val := range test.ListVal {
				if err == nil && val.Validate() != nil {
					t.Fatal(val.Validate())
				}
			}
		})
		t.Run(test.Name+"_authorize", func(t *testing.T) {
			err := test.Client.Authorize(test.Ctx, test.AuthVerb, test.AuthName)
			if err != nil && !test.AuthErr {
				t.Fatal(err)
			}
			if err == nil && test.AuthErr {
				t.Fatal(err)
			}
		})
	}
}

func TestSetTypeMeta(t *testing.T) {
	var g GenericClient
	err := g.SetTypeMeta(corev2.TypeMeta{
		Type:       "check",
		APIVersion: "core/v2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := g.Kind, (&corev2.Check{}); !reflect.DeepEqual(got, want) {
		t.Fatal("SetTypeMeta not working")
	}
}

func TestGenericClient_SetTypeMeta(t *testing.T) {
	tests := []struct {
		name       string
		arg        corev2.TypeMeta
		outGroup   string
		outVersion string
		wantErr    bool
	}{
		{
			name: "unregistered type",
			arg: corev2.TypeMeta{
				APIVersion: "core/v2",
				Type:       "xxxyyyzzz",
			},
			outGroup:   "core",
			outVersion: "v2",
			wantErr:    true,
		},
		{
			name: "registered type",
			arg: corev2.TypeMeta{
				APIVersion: "core/v2",
				Type:       "mutator",
			},
			outGroup:   "core",
			outVersion: "v2",
			wantErr:    false,
		},
		{
			name: "defaults to core/v2",
			arg: corev2.TypeMeta{
				Type: "mutator",
			},
			outGroup:   "core",
			outVersion: "v2",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GenericClient{}
			if err := g.SetTypeMeta(tt.arg); (err != nil) != tt.wantErr {
				t.Errorf("GenericClient.SetTypeMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.outGroup != g.APIGroup {
				t.Errorf("GenericClient.SetTypeMeta() group = %v, expected group = %v", g.APIGroup, tt.outGroup)
			}
			if tt.outVersion != g.APIVersion {
				t.Errorf("GenericClient.SetTypeMeta() version = %v, expected version = %v", g.APIVersion, tt.outVersion)
			}
		})
	}
}

func TestGenericClientStoreV2(t *testing.T) {
	tests := []struct {
		Name      string
		Client    *GenericClient
		CreateVal corev3.Resource
		CreateErr bool
		UpdateVal corev3.Resource
		UpdateErr bool
		GetName   string
		GetVal    corev3.Resource
		GetErr    bool
		ListVal   []corev3.Resource
		ListPred  *store.SelectionPredicate
		ListErr   bool
		DelName   string
		DelErr    bool
		Ctx       context.Context
	}{
		{
			Name:      "authorizer rejects all",
			Client:    defaultV2TestClient(defaultV2ResourceStore(), badAuth()),
			CreateVal: defaultV3Resource(),
			CreateErr: true,
			UpdateVal: defaultV3Resource(),
			UpdateErr: true,
			GetName:   "toget",
			GetVal:    defaultV3Resource(),
			GetErr:    true,
			DelName:   "todelete",
			DelErr:    true,
			ListErr:   true,
			Ctx:       contextWithUser(defaultContext(), "tom", nil),
		},
		{
			Name: "readonly access",
			Client: defaultV2TestClient(defaultV2ResourceStore(), &mockAuth{
				attrs: map[authorization.AttributesKey]bool{
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v3",
						Namespace:    "default",
						Resource:     "entity_configs",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "get",
					}: true,
					authorization.AttributesKey{
						APIGroup:   "core",
						APIVersion: "v3",
						Namespace:  "default",
						Resource:   "entity_configs",
						UserName:   "tom",
						Verb:       "list",
					}: true,
				},
			}),
			CreateVal: defaultV3Resource(),
			CreateErr: true,
			UpdateVal: defaultV3Resource(),
			UpdateErr: true,
			GetName:   "default",
			GetVal:    defaultV3Resource(),
			GetErr:    false,
			DelName:   "default",
			DelErr:    true,
			ListErr:   false,
			Ctx:       contextWithUser(defaultContext(), "tom", nil),
		},
		{
			Name: "all access",
			Client: defaultV2TestClient(defaultV2ResourceStore(), &mockAuth{
				attrs: map[authorization.AttributesKey]bool{
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v3",
						Namespace:    "default",
						Resource:     "entity_configs",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "get",
					}: true,
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v3",
						Namespace:    "default",
						Resource:     "entity_configs",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "create",
					}: true,
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v3",
						Namespace:    "default",
						Resource:     "entity_configs",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "delete",
					}: true,
					authorization.AttributesKey{
						APIGroup:     "core",
						APIVersion:   "v3",
						Namespace:    "default",
						Resource:     "entity_configs",
						ResourceName: "default",
						UserName:     "tom",
						Verb:         "update",
					}: true,
					authorization.AttributesKey{
						APIGroup:   "core",
						APIVersion: "v3",
						Namespace:  "default",
						Resource:   "entity_configs",
						UserName:   "tom",
						Verb:       "list",
					}: true,
				},
			}),
			CreateVal: defaultV3Resource(),
			CreateErr: false,
			UpdateVal: defaultV3Resource(),
			UpdateErr: false,
			GetName:   "default",
			GetVal:    defaultV3Resource(),
			GetErr:    false,
			DelName:   "default",
			DelErr:    false,
			ListErr:   false,
			Ctx:       contextWithUser(defaultContext(), "tom", nil),
		},
	}

	for _, test := range tests {
		t.Run(test.Name+"_create", func(t *testing.T) {
			err := test.Client.Create(test.Ctx, test.CreateVal)
			if err != nil && !test.CreateErr {
				t.Fatal(err)
			}
			if err == nil && test.CreateErr {
				t.Fatal("expected non-nil error")
			}
		})
		t.Run(test.Name+"_update", func(t *testing.T) {
			err := test.Client.Update(test.Ctx, test.UpdateVal)
			if err != nil && !test.UpdateErr {
				t.Fatal(err)
			}
			if err == nil && test.UpdateErr {
				t.Fatal(err)
			}
		})
		t.Run(test.Name+"_get", func(t *testing.T) {
			err := test.Client.Get(test.Ctx, test.GetName, test.GetVal)
			if err != nil && !test.GetErr {
				t.Fatal(err)
			}
			if err == nil && test.GetErr {
				t.Fatal(err)
			}
			if err == nil && test.GetVal.Validate() != nil {
				t.Fatal(test.GetVal.Validate())
			}
		})
		t.Run(test.Name+"_del", func(t *testing.T) {
			err := test.Client.Delete(test.Ctx, test.DelName)
			if err != nil && !test.DelErr {
				t.Fatal(err)
			}
			if err == nil && test.DelErr {
				t.Fatal(err)
			}
		})
		t.Run(test.Name+"_list", func(t *testing.T) {
			err := test.Client.List(test.Ctx, &test.ListVal, test.ListPred)
			if err != nil && !test.ListErr {
				t.Fatal(err)
			}
			if err == nil && test.ListErr {
				t.Fatal(err)
			}
			for _, val := range test.ListVal {
				if err == nil && val.Validate() != nil {
					t.Fatal(val.Validate())
				}
			}
		})
	}
}

func TestSetTypeMetaV3Resource(t *testing.T) {
	client := &GenericClient{}
	err := client.SetTypeMeta(corev2.TypeMeta{
		APIVersion: "core/v3",
		Type:       "EntityConfig",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := client.Kind.(*corev3.EntityConfig); !ok {
		t.Error("expected an entityconfig")
	}
}
