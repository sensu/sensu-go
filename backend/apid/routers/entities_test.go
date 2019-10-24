package routers

import (
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func TestEntitiesRouter(t *testing.T) {
	// Setup the router
	s := &mockstore.MockStore{}
	s.On("GetEventsByEntity", mock.Anything, "foo", mock.Anything).Return([]*corev2.Event{corev2.FixtureEvent("foo", "bar")}, nil)
	s.On("DeleteEventByEntityCheck", mock.Anything, "foo", "bar").Return(nil)
	s.On("DeleteEntityByName", mock.Anything, "foo").Return(nil)
	router := NewEntitiesRouter(s, s)
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
	router.Mount(parentRouter)

	empty := &corev2.Entity{}
	fixture := corev2.FixtureEntity("foo")

	tests := []routerTestCase{}
	tests = append(tests, getTestCases(fixture)...)
	tests = append(tests, listTestCases(empty)...)
	// NB: we avoid some of the generic deletion tests here since they
	// are incompatible with the specialized deletion logic of the entity
	// controller.
	tests = append(tests, deleteResourceInvalidPathTestCase(fixture))
	tests = append(tests, deleteResourceSuccessTestCase(fixture))
	for _, tt := range tests {
		run(t, tt, parentRouter, s)
	}
}

func TestEntitiesCreateRouter(t *testing.T) {
	// Setup the router
	s := &mockstore.MockStore{}
	s.On("GetEventsByEntity", mock.Anything, "foo", mock.Anything).Return([]*corev2.Event{corev2.FixtureEvent("foo", "bar")}, nil)
	s.On("DeleteEventByEntityCheck", mock.Anything, "foo", "bar").Return(nil)
	s.On("DeleteEntityByName", mock.Anything, "foo").Return(nil)
	router := NewEntitiesCreateRouter(s, s)
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
	router.Mount(parentRouter)

	empty := &corev2.Entity{}
	fixture := corev2.FixtureEntity("foo")

	tests := []routerTestCase{}
	tests = append(tests, createTestCases(empty)...)
	tests = append(tests, updateTestCases(fixture)...)
	for _, tt := range tests {
		run(t, tt, parentRouter, s)
	}
}
