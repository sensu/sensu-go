package routers

import (
	"context"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

type mockEntitiesController struct {
	mock.Mock
}

func (m *mockEntitiesController) Find(ctx context.Context, id string) (*corev2.Entity, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*corev2.Entity), args.Error(1)
}

func (m *mockEntitiesController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	args := m.Called(ctx, pred)
	return args.Get(0).([]corev2.Resource), args.Error(1)
}

func (m *mockEntitiesController) Create(ctx context.Context, entity corev2.Entity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *mockEntitiesController) CreateOrReplace(ctx context.Context, entity corev2.Entity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func TestEntitiesRouter(t *testing.T) {
	// Setup the router
	controller := new(mockEntitiesController)
	controller.On("Find", mock.Anything, mock.Anything).Return(corev2.FixtureEntity("foo"), nil)
	controller.On("List", mock.Anything, mock.Anything).Return([]corev2.Resource{corev2.FixtureEntity("foo")}, nil)
	controller.On("Create", mock.Anything, mock.Anything).Return(nil)
	controller.On("CreateOrReplace", mock.Anything, mock.Anything).Return(nil)
	s := new(mockstore.MockStore)
	s.On("GetEventsByEntity", mock.Anything, "foo", mock.Anything).Return([]*corev2.Event{corev2.FixtureEvent("foo", "bar")}, nil)
	s.On("DeleteEventByEntityCheck", mock.Anything, "foo", "bar").Return(nil)
	s.On("DeleteEntityByName", mock.Anything, "foo").Return(nil)
	s.On("GetEntityByName", mock.Anything, "foo").Return(corev2.FixtureEntity("foo"), nil)
	router := NewEntitiesRouter(s, s)
	router.controller = controller
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
	router.Mount(parentRouter)

	//empty := &corev2.Entity{}
	fixture := corev2.FixtureEntity("foo")

	tests := []routerTestCase{}
	//TODO(eric): replace these test cases so they work without handlers
	//tests = append(tests, getTestCases(fixture)...)
	//tests = append(tests, listTestCases(empty)...)
	//tests = append(tests, createTestCases(empty)...)
	//tests = append(tests, updateTestCases(fixture)...)
	// NB: we avoid some of the generic deletion tests here since they
	// are incompatible with the specialized deletion logic of the entity
	// controller.
	tests = append(tests, deleteResourceInvalidPathTestCase(fixture))
	tests = append(tests, deleteResourceSuccessTestCase(fixture))
	for _, tt := range tests {
		run(t, tt, parentRouter, s)
	}
}
