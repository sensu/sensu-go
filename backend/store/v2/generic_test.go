package v2_test

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

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

func TestGenericStorePatch(t *testing.T) {
	sv2 := new(mockstore.V2MockStore)
	sv2.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	store := storev2.NewGenericStore[*corev2.CheckConfig](sv2)
	if err := store.Patch(context.Background(), corev2.FixtureCheckConfig("foo"), nil, nil); err != nil {
		t.Fatal(err)
	}
	sv2.AssertCalled(t, "Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}
