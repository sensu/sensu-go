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

type mockSilencedController struct {
	mock.Mock
}

func (m *mockSilencedController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	args := m.Called(ctx, pred)
	return args.Get(0).([]corev2.Resource), args.Error(1)
}

func TestSilencedRouter(t *testing.T) {
	// Setup the router
	s := &mockstore.MockStore{}
	router := NewSilencedRouter(s)
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
	router.Mount(parentRouter)

	fixture := corev2.FixtureSilenced("foo:bar")

	tests := []routerTestCase{}
	tests = append(tests, getTestCases(fixture)...)
	tests = append(tests, deleteTestCases(fixture)...)
	// TODO(palourde): Re-enable these tests once the silenced router uses the
	// common handlers
	// tests = append(tests, listTestCases(pathPrefix, kind, []corev2.Resource{fixture})...)
	// tests = append(tests, createTestCases(pathPrefix, kind)...)
	// tests = append(tests, updateTestCases(pathPrefix, kind)...)
	for _, tt := range tests {
		run(t, tt, parentRouter, s)
	}
}
