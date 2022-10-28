package v2_test

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

type extendedMockStoreV2 struct {
	mockstore.V2MockStore
}

func (e *extendedMockStoreV2) GetEntityConfigStore() storev2.EntityConfigStore {
	return e.Called().Get(0).(storev2.EntityConfigStore)
}

type mockEntityConfigStore struct {
	mock.Mock
}

func (m *mockEntityConfigStore) CreateOrUpdate(ctx context.Context, entity *corev3.EntityConfig) error {
	return m.Called(ctx, entity).Error(0)
}
func (m *mockEntityConfigStore) UpdateIfExists(ctx context.Context, entity *corev3.EntityConfig) error {
	return m.Called(ctx, entity).Error(0)
}
func (m *mockEntityConfigStore) CreateIfNotExists(ctx context.Context, entity *corev3.EntityConfig) error {
	return m.Called(ctx, entity).Error(0)
}
func (m *mockEntityConfigStore) Get(ctx context.Context, namespace string, name string) (*corev3.EntityConfig, error) {
	args := m.Called(ctx, namespace, name)
	return args.Get(0).(*corev3.EntityConfig), args.Error(1)
}
func (m *mockEntityConfigStore) Delete(ctx context.Context, namespace string, name string) error {
	return m.Called(ctx, namespace, name).Error(0)
}
func (m *mockEntityConfigStore) List(ctx context.Context, namespace string, pred *store.SelectionPredicate) ([]*corev3.EntityConfig, error) {
	args := m.Called(ctx, namespace, pred)
	return args.Get(0).([]*corev3.EntityConfig), args.Error(1)
}
func (m *mockEntityConfigStore) Count(ctx context.Context, namespace string, entityClass string) (int, error) {
	args := m.Called(ctx, namespace, entityClass)
	return args.Get(0).(int), args.Error(1)
}
func (m *mockEntityConfigStore) Exists(ctx context.Context, namespace string, name string) (bool, error) {
	args := m.Called(ctx, namespace, name)
	return args.Get(0).(bool), args.Error(1)
}
func (m *mockEntityConfigStore) Patch(ctx context.Context, namespace string, name string, patcher patch.Patcher, cond *store.ETagCondition) error {
	return m.Called(ctx, namespace, name, patcher, cond).Error(0)
}

func (e *extendedMockStoreV2) GetEntityStateStore() storev2.EntityStateStore {
	return e.Called().Get(0).(storev2.EntityStateStore)
}

type mockEntityStateStore struct {
	mock.Mock
}

func (m *mockEntityStateStore) CreateOrUpdate(ctx context.Context, entity *corev3.EntityState) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *mockEntityStateStore) UpdateIfExists(ctx context.Context, entity *corev3.EntityState) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *mockEntityStateStore) CreateIfNotExists(ctx context.Context, entity *corev3.EntityState) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *mockEntityStateStore) Get(ctx context.Context, namespace string, name string) (*corev3.EntityState, error) {
	args := m.Called(ctx, namespace, name)
	return args.Get(0).(*corev3.EntityState), args.Error(1)
}

func (m *mockEntityStateStore) Delete(ctx context.Context, namespace string, name string) error {
	return m.Called(ctx, namespace, name).Error(0)
}

func (m *mockEntityStateStore) List(ctx context.Context, namespace string, pred *store.SelectionPredicate) ([]*corev3.EntityState, error) {
	args := m.Called(ctx, namespace, pred)
	return args.Get(0).([]*corev3.EntityState), args.Error(1)
}

func (m *mockEntityStateStore) Count(ctx context.Context, namespace string) (int, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(int), args.Error(1)
}
func (m *mockEntityStateStore) Exists(ctx context.Context, namespace string, name string) (bool, error) {
	args := m.Called(ctx, namespace, name)
	return args.Get(0).(bool), args.Error(1)
}

func (m *mockEntityStateStore) Patch(ctx context.Context, namespace string, name string, patcher patch.Patcher, cond *store.ETagCondition) error {
	return m.Called(ctx, namespace, name, patcher, cond).Error(0)
}

func (e *extendedMockStoreV2) GetNamespaceStore() storev2.NamespaceStore {
	return e.Called().Get(0).(storev2.NamespaceStore)
}

type mockNamespaceStore struct {
	mock.Mock
}

func (m *mockNamespaceStore) CreateOrUpdate(ctx context.Context, namespace *corev3.Namespace) error {
	return m.Called(ctx, namespace).Error(0)
}

func (m *mockNamespaceStore) UpdateIfExists(ctx context.Context, namespace *corev3.Namespace) error {
	return m.Called(ctx, namespace).Error(0)
}

func (m *mockNamespaceStore) CreateIfNotExists(ctx context.Context, namespace *corev3.Namespace) error {
	return m.Called(ctx, namespace).Error(0)
}

func (m *mockNamespaceStore) Get(ctx context.Context, namespace string) (*corev3.Namespace, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(*corev3.Namespace), args.Error(1)
}

func (m *mockNamespaceStore) Delete(ctx context.Context, namespace string) error {
	return m.Called(ctx, namespace).Error(0)
}

func (m *mockNamespaceStore) List(ctx context.Context, pred *store.SelectionPredicate) ([]*corev3.Namespace, error) {
	args := m.Called(ctx, pred)
	return args.Get(0).([]*corev3.Namespace), args.Error(1)
}

func (m *mockNamespaceStore) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

func (m *mockNamespaceStore) Exists(ctx context.Context, namespace string) (bool, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(bool), args.Error(1)
}

func (m *mockNamespaceStore) IsEmpty(ctx context.Context, namespace string) (bool, error) {
	args := m.Called(ctx, namespace)
	return args.Get(0).(bool), args.Error(1)
}

func (m *mockNamespaceStore) Patch(ctx context.Context, namespace string, patcher patch.Patcher, cond *store.ETagCondition) error {
	return m.Called(ctx, namespace, patcher, cond).Error(0)
}

func TestGenericStoreCreateOrUpdate(t *testing.T) {
	sv2 := new(mockstore.V2MockStore)
	req := storev2.ResourceRequest{
		APIVersion: "core/v2",
		Type:       "CheckConfig",
		Namespace:  "default",
		Name:       "foo",
		StoreName:  "check_configs",
	}
	sv2.On("CreateOrUpdate", mock.Anything, req, mock.Anything).Return(nil)
	store := storev2.NewGenericStore[*corev2.CheckConfig](sv2)
	if err := store.CreateOrUpdate(context.Background(), corev2.FixtureCheckConfig("foo")); err != nil {
		t.Fatal(err)
	}
	sv2.AssertCalled(t, "CreateOrUpdate", mock.Anything, req, mock.Anything)
}

func TestGenericStoreCreateOrUpdateEntityConfig(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entConfigStore := new(mockEntityConfigStore)
	entConfigStore.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetEntityConfigStore").Return(entConfigStore)
	store := storev2.NewGenericStore[*corev3.EntityConfig](sv2)
	if err := store.CreateOrUpdate(context.Background(), corev3.FixtureEntityConfig("foo")); err != nil {
		t.Fatal(err)
	}
	entConfigStore.AssertCalled(t, "CreateOrUpdate", mock.Anything, mock.Anything)
}

func TestGenericStoreCreateOrUpdateEntityState(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entStateStore := new(mockEntityStateStore)
	entStateStore.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetEntityStateStore").Return(entStateStore)
	store := storev2.NewGenericStore[*corev3.EntityState](sv2)
	if err := store.CreateOrUpdate(context.Background(), corev3.FixtureEntityState("foo")); err != nil {
		t.Fatal(err)
	}
	entStateStore.AssertCalled(t, "CreateOrUpdate", mock.Anything, mock.Anything)
}

func TestGenericStoreCreateOrUpdateNamespace(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	nsStore := new(mockNamespaceStore)
	nsStore.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetNamespaceStore").Return(nsStore)
	store := storev2.NewGenericStore[*corev3.Namespace](sv2)
	if err := store.CreateOrUpdate(context.Background(), corev3.FixtureNamespace("foo")); err != nil {
		t.Fatal(err)
	}
	nsStore.AssertCalled(t, "CreateOrUpdate", mock.Anything, mock.Anything)
}

func TestGenericStoreUpdateIfExists(t *testing.T) {
	sv2 := new(mockstore.V2MockStore)
	req := storev2.ResourceRequest{
		APIVersion: "core/v2",
		Type:       "CheckConfig",
		Namespace:  "default",
		Name:       "foo",
		StoreName:  "check_configs",
	}
	sv2.On("UpdateIfExists", mock.Anything, req, mock.Anything).Return(nil)
	store := storev2.NewGenericStore[*corev2.CheckConfig](sv2)
	if err := store.UpdateIfExists(context.Background(), corev2.FixtureCheckConfig("foo")); err != nil {
		t.Fatal(err)
	}
	sv2.AssertCalled(t, "UpdateIfExists", mock.Anything, req, mock.Anything)
}

func TestGenericStoreUpdateIfExistsEntityConfig(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entConfigStore := new(mockEntityConfigStore)
	entConfigStore.On("UpdateIfExists", mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetEntityConfigStore").Return(entConfigStore)
	store := storev2.NewGenericStore[*corev3.EntityConfig](sv2)
	if err := store.UpdateIfExists(context.Background(), corev3.FixtureEntityConfig("foo")); err != nil {
		t.Fatal(err)
	}
	entConfigStore.AssertCalled(t, "UpdateIfExists", mock.Anything, mock.Anything)
}

func TestGenericStoreUpdateIfExistsEntityState(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entStateStore := new(mockEntityStateStore)
	entStateStore.On("UpdateIfExists", mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetEntityStateStore").Return(entStateStore)
	store := storev2.NewGenericStore[*corev3.EntityState](sv2)
	if err := store.UpdateIfExists(context.Background(), corev3.FixtureEntityState("foo")); err != nil {
		t.Fatal(err)
	}
	entStateStore.AssertCalled(t, "UpdateIfExists", mock.Anything, mock.Anything)
}

func TestGenericStoreUpdateIfExistsNamespace(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	nsStore := new(mockNamespaceStore)
	nsStore.On("UpdateIfExists", mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetNamespaceStore").Return(nsStore)
	store := storev2.NewGenericStore[*corev3.Namespace](sv2)
	if err := store.UpdateIfExists(context.Background(), corev3.FixtureNamespace("foo")); err != nil {
		t.Fatal(err)
	}
	nsStore.AssertCalled(t, "UpdateIfExists", mock.Anything, mock.Anything)
}

func TestGenericStoreCreateIfNotExists(t *testing.T) {
	sv2 := new(mockstore.V2MockStore)
	req := storev2.ResourceRequest{
		APIVersion: "core/v2",
		Type:       "CheckConfig",
		Namespace:  "default",
		Name:       "foo",
		StoreName:  "check_configs",
	}
	sv2.On("CreateIfNotExists", mock.Anything, req, mock.Anything).Return(nil)
	store := storev2.NewGenericStore[*corev2.CheckConfig](sv2)
	if err := store.CreateIfNotExists(context.Background(), corev2.FixtureCheckConfig("foo")); err != nil {
		t.Fatal(err)
	}
	sv2.AssertCalled(t, "CreateIfNotExists", mock.Anything, req, mock.Anything)
}

func TestGenericStoreCreateIfNotExistsEntityConfig(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entConfigStore := new(mockEntityConfigStore)
	entConfigStore.On("CreateIfNotExists", mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetEntityConfigStore").Return(entConfigStore)
	store := storev2.NewGenericStore[*corev3.EntityConfig](sv2)
	if err := store.CreateIfNotExists(context.Background(), corev3.FixtureEntityConfig("foo")); err != nil {
		t.Fatal(err)
	}
	entConfigStore.AssertCalled(t, "CreateIfNotExists", mock.Anything, mock.Anything)
}

func TestGenericStoreCreateIfNotExistsEntityState(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entStateStore := new(mockEntityStateStore)
	entStateStore.On("CreateIfNotExists", mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetEntityStateStore").Return(entStateStore)
	store := storev2.NewGenericStore[*corev3.EntityState](sv2)
	if err := store.CreateIfNotExists(context.Background(), corev3.FixtureEntityState("foo")); err != nil {
		t.Fatal(err)
	}
	entStateStore.AssertCalled(t, "CreateIfNotExists", mock.Anything, mock.Anything)
}

func TestGenericStoreCreateIfNotExistsNamespace(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	nsStore := new(mockNamespaceStore)
	nsStore.On("CreateIfNotExists", mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetNamespaceStore").Return(nsStore)
	store := storev2.NewGenericStore[*corev3.Namespace](sv2)
	if err := store.CreateIfNotExists(context.Background(), corev3.FixtureNamespace("foo")); err != nil {
		t.Fatal(err)
	}
	nsStore.AssertCalled(t, "CreateIfNotExists", mock.Anything, mock.Anything)
}

func TestGenericStoreGet(t *testing.T) {
	sv2 := new(mockstore.V2MockStore)
	req := storev2.ResourceRequest{
		APIVersion: "core/v2",
		Type:       "CheckConfig",
		Namespace:  "default",
		Name:       "foo",
		StoreName:  "check_configs",
	}
	sv2.On("Get", mock.Anything, req).Return(mockstore.Wrapper[*corev2.CheckConfig]{Value: corev2.FixtureCheckConfig("foo")}, nil)
	store := storev2.NewGenericStore[*corev2.CheckConfig](sv2)
	check, err := store.Get(context.Background(), storev2.ID{Namespace: "default", Name: "foo"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := check.Name, "foo"; got != want {
		t.Fatalf("wrong check name: got %q, want %q", got, want)
	}
	if got, want := check.Interval, uint32(60); got != want {
		t.Fatalf("wrong check interval: got %d, want %d", got, want)
	}
	sv2.AssertCalled(t, "Get", mock.Anything, req)
}

func TestGenericStoreGetEntityConfig(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entConfigStore := new(mockEntityConfigStore)
	entConfigStore.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(corev3.FixtureEntityConfig("foo"), nil)
	sv2.On("GetEntityConfigStore").Return(entConfigStore)
	store := storev2.NewGenericStore[*corev3.EntityConfig](sv2)
	if _, err := store.Get(context.Background(), storev2.ID{}); err != nil {
		t.Fatal(err)
	}
	entConfigStore.AssertCalled(t, "Get", mock.Anything, mock.Anything, mock.Anything)
}

func TestGenericStoreGetEntityState(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entStateStore := new(mockEntityStateStore)
	entStateStore.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(corev3.FixtureEntityState("foo"), nil)
	sv2.On("GetEntityStateStore").Return(entStateStore)
	store := storev2.NewGenericStore[*corev3.EntityState](sv2)
	if _, err := store.Get(context.Background(), storev2.ID{}); err != nil {
		t.Fatal(err)
	}
	entStateStore.AssertCalled(t, "Get", mock.Anything, mock.Anything, mock.Anything)
}

func TestGenericStoreGetNamespace(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	nsStore := new(mockNamespaceStore)
	nsStore.On("Get", mock.Anything, mock.Anything).Return(corev3.FixtureNamespace("foo"), nil)
	sv2.On("GetNamespaceStore").Return(nsStore)
	store := storev2.NewGenericStore[*corev3.Namespace](sv2)
	if _, err := store.Get(context.Background(), storev2.ID{}); err != nil {
		t.Fatal(err)
	}
	nsStore.AssertCalled(t, "Get", mock.Anything, mock.Anything)
}

func TestGenericStoreDelete(t *testing.T) {
	sv2 := new(mockstore.V2MockStore)
	req := storev2.ResourceRequest{
		APIVersion: "core/v2",
		Type:       "CheckConfig",
		Namespace:  "default",
		Name:       "foo",
		StoreName:  "check_configs",
	}
	sv2.On("Delete", mock.Anything, req).Return(nil)
	store := storev2.NewGenericStore[*corev2.CheckConfig](sv2)
	if err := store.Delete(context.Background(), storev2.ID{Namespace: "default", Name: "foo"}); err != nil {
		t.Fatal(err)
	}
	sv2.AssertCalled(t, "Delete", mock.Anything, req)
}

func TestGenericStoreDeleteEntityConfig(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entConfigStore := new(mockEntityConfigStore)
	entConfigStore.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetEntityConfigStore").Return(entConfigStore)
	store := storev2.NewGenericStore[*corev3.EntityConfig](sv2)
	if err := store.Delete(context.Background(), storev2.ID{}); err != nil {
		t.Fatal(err)
	}
	entConfigStore.AssertCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
}

func TestGenericStoreDeleteEntityState(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entStateStore := new(mockEntityStateStore)
	entStateStore.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetEntityStateStore").Return(entStateStore)
	store := storev2.NewGenericStore[*corev3.EntityState](sv2)
	if err := store.Delete(context.Background(), storev2.ID{}); err != nil {
		t.Fatal(err)
	}
	entStateStore.AssertCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
}

func TestGenericStoreDeleteNamespace(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	nsStore := new(mockNamespaceStore)
	nsStore.On("Delete", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetNamespaceStore").Return(nsStore)
	store := storev2.NewGenericStore[*corev3.Namespace](sv2)
	if err := store.Delete(context.Background(), storev2.ID{}); err != nil {
		t.Fatal(err)
	}
	nsStore.AssertCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestGenericStoreList(t *testing.T) {
	sv2 := new(mockstore.V2MockStore)
	req := storev2.ResourceRequest{
		APIVersion: "core/v2",
		Type:       "CheckConfig",
		Namespace:  "default",
		Name:       "",
		StoreName:  "check_configs",
	}
	wrapper, _ := wrap.Resource(corev2.FixtureCheckConfig("foo"))
	wrapList := wrap.List{wrapper}
	sv2.On("List", mock.Anything, req, mock.Anything).Return(wrapList, nil)
	store := storev2.NewGenericStore[*corev2.CheckConfig](sv2)
	checks, err := store.List(context.Background(), storev2.ID{Namespace: "default", Name: "foo"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := checks[0].Name, "foo"; got != want {
		t.Fatalf("wrong check name: got %q, want %q", got, want)
	}
	if got, want := checks[0].Interval, uint32(60); got != want {
		t.Fatalf("wrong check interval: got %d, want %d", got, want)
	}
	sv2.AssertCalled(t, "List", mock.Anything, req, mock.Anything)
}

func TestGenericStoreListEntityConfig(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entConfigStore := new(mockEntityConfigStore)
	entConfigStore.On("List", mock.Anything, mock.Anything, mock.Anything).Return([]*corev3.EntityConfig{}, nil)
	sv2.On("GetEntityConfigStore").Return(entConfigStore)
	store := storev2.NewGenericStore[*corev3.EntityConfig](sv2)
	if _, err := store.List(context.Background(), storev2.ID{}, nil); err != nil {
		t.Fatal(err)
	}
	entConfigStore.AssertCalled(t, "List", mock.Anything, mock.Anything, mock.Anything)
}

func TestGenericStoreListEntityState(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entStateStore := new(mockEntityStateStore)
	entStateStore.On("List", mock.Anything, mock.Anything, mock.Anything).Return([]*corev3.EntityState{}, nil)
	sv2.On("GetEntityStateStore").Return(entStateStore)
	store := storev2.NewGenericStore[*corev3.EntityState](sv2)
	if _, err := store.List(context.Background(), storev2.ID{}, nil); err != nil {
		t.Fatal(err)
	}
	entStateStore.AssertCalled(t, "List", mock.Anything, mock.Anything, mock.Anything)
}

func TestGenericStoreListNamespace(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	nsStore := new(mockNamespaceStore)
	nsStore.On("List", mock.Anything, mock.Anything).Return([]*corev3.Namespace{}, nil)
	sv2.On("GetNamespaceStore").Return(nsStore)
	store := storev2.NewGenericStore[*corev3.Namespace](sv2)
	if _, err := store.List(context.Background(), storev2.ID{}, nil); err != nil {
		t.Fatal(err)
	}
	nsStore.AssertCalled(t, "List", mock.Anything, mock.Anything)
}

func TestGenericStorePatch(t *testing.T) {
	sv2 := new(mockstore.V2MockStore)
	sv2.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	store := storev2.NewGenericStore[*corev2.CheckConfig](sv2)
	if err := store.Patch(context.Background(), corev2.FixtureCheckConfig("foo"), nil, nil); err != nil {
		t.Fatal(err)
	}
	sv2.AssertCalled(t, "Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGenericStorePatchEntityConfig(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entConfigStore := new(mockEntityConfigStore)
	entConfigStore.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetEntityConfigStore").Return(entConfigStore)
	store := storev2.NewGenericStore[*corev3.EntityConfig](sv2)
	if err := store.Patch(context.Background(), corev3.FixtureEntityConfig("foo"), nil, nil); err != nil {
		t.Fatal(err)
	}
	entConfigStore.AssertCalled(t, "Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGenericStorePatchEntityState(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	entStateStore := new(mockEntityStateStore)
	entStateStore.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetEntityStateStore").Return(entStateStore)
	store := storev2.NewGenericStore[*corev3.EntityState](sv2)
	if err := store.Patch(context.Background(), corev3.FixtureEntityState("foo"), nil, nil); err != nil {
		t.Fatal(err)
	}
	entStateStore.AssertCalled(t, "Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestGenericStorePatchNamespace(t *testing.T) {
	sv2 := new(extendedMockStoreV2)
	nsStore := new(mockNamespaceStore)
	nsStore.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	sv2.On("GetNamespaceStore").Return(nsStore)
	store := storev2.NewGenericStore[*corev3.Namespace](sv2)
	if err := store.Patch(context.Background(), corev3.FixtureNamespace("foo"), nil, nil); err != nil {
		t.Fatal(err)
	}
	nsStore.AssertCalled(t, "Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
