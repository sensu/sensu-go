package routers

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

func TestSilencedRouter(t *testing.T) {
	// Setup the router
	s := &mockstore.V2MockStore{}
	cs := new(mockstore.ConfigStore)
	s.On("GetConfigStore").Return(cs)
	silenced := &mockstore.MockStore{}
	s.On("GetSilencesStore").Return(silenced)
	router := NewSilencedRouter(s)
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
	router.Mount(parentRouter)

	fixture := corev2.FixtureSilenced("*:bar")

	tests := []routerTestCase{}
	tests = append(tests, deleteTestCases(fixture)...)
	for _, tt := range tests {
		run(t, tt, parentRouter, s)
	}
}

type mockSilencedController struct {
	mock.Mock
}

func (m *mockSilencedController) Create(ctx context.Context, entry *corev2.Silenced) error {
	return m.Called(ctx, entry).Error(0)
}

func (m *mockSilencedController) CreateOrReplace(ctx context.Context, entry *corev2.Silenced) error {
	return m.Called(ctx, entry).Error(0)
}

func (m *mockSilencedController) List(ctx context.Context, sub, check string) ([]*corev2.Silenced, error) {
	args := m.Called(ctx, sub, check)
	return args.Get(0).([]*corev2.Silenced), args.Error(1)
}

func (m *mockSilencedController) Get(ctx context.Context, name string) (*corev2.Silenced, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*corev2.Silenced), args.Error(1)
}

func TestSilencedRouterCustomRoutes(t *testing.T) {
	type controllerFunc func(*mockSilencedController)

	// Setup the router
	controller := &mockSilencedController{}
	router := SilencedRouter{controller: controller}
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
	router.Mount(parentRouter)

	empty := &corev2.Silenced{}
	empty.SetNamespace("default")
	fixture := corev2.FixtureSilenced("linux:check-cpu")

	tests := []struct {
		name           string
		method         string
		path           string
		body           []byte
		controllerFunc controllerFunc
		wantStatusCode int
	}{
		{
			name:           "it returns 400 if the payload to create is not decodable",
			method:         http.MethodPost,
			path:           empty.URIPath(),
			body:           []byte(`foo`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "it returns 400 if the silenced entry metadata to create is invalid",
			method:         http.MethodPost,
			path:           empty.URIPath(),
			body:           []byte(`{"metadata": {"namespace":"acme"}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "it returns 500 if the store returns an error while creating a silenced entry",
			method: http.MethodPost,
			path:   empty.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockSilencedController) {
				c.On("Create", mock.Anything, mock.Anything).
					Return(actions.NewErrorf(actions.InternalErr)).
					Once()
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "it returns 201 when an event is successfully created",
			method: http.MethodPost,
			path:   empty.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockSilencedController) {
				c.On("Create", mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "it returns 400 if the payload to update is not decodable",
			method:         http.MethodPut,
			path:           fixture.URIPath(),
			body:           []byte(`foo`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "it returns 400 if the silenced entry metadata to update is invalid",
			method:         http.MethodPut,
			path:           fixture.URIPath(),
			body:           []byte(`{"metadata": {"namespace":"acme"}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "it returns 500 if the store returns an error while updating a silenced entry",
			method: http.MethodPut,
			path:   fixture.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockSilencedController) {
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
					Return(actions.NewErrorf(actions.InternalErr)).
					Once()
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "it returns 201 when a silenced entry is successfully updated",
			method: http.MethodPut,
			path:   fixture.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockSilencedController) {
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:   "it returns 200 on successful get",
			method: http.MethodGet,
			path:   corev2.FixtureSilenced("*:foo").URIPath(),
			body:   nil,
			controllerFunc: func(c *mockSilencedController) {
				c.On("Get", mock.Anything, "*:foo").
					Return(corev2.FixtureSilenced(":foo"), nil).
					Once()
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "it returns 404 on unsuccessful get",
			method: http.MethodGet,
			path:   corev2.FixtureSilenced("not:exists").URIPath(),
			body:   nil,
			controllerFunc: func(c *mockSilencedController) {
				c.On("Get", mock.Anything, mock.Anything).
					Return((*corev2.Silenced)(nil), actions.NewError(actions.NotFound, errors.New("not found"))).
					Once()
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:   "it returns 500 on get error",
			method: http.MethodGet,
			path:   corev2.FixtureSilenced("not:exists").URIPath(),
			body:   nil,
			controllerFunc: func(c *mockSilencedController) {
				c.On("Get", mock.Anything, mock.Anything).
					Return((*corev2.Silenced)(nil), errors.New("hi")).
					Once()
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "it returns 200 on successful list",
			method: http.MethodGet,
			path:   new(corev2.Silenced).URIPath(),
			body:   nil,
			controllerFunc: func(c *mockSilencedController) {
				c.On("List", mock.Anything, "", "").
					Return([]*corev2.Silenced{}, nil).
					Once()
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "it returns 500 on error list",
			method: http.MethodGet,
			path:   new(corev2.Silenced).URIPath(),
			body:   nil,
			controllerFunc: func(c *mockSilencedController) {
				c.On("List", mock.Anything, "", "").
					Return(([]*corev2.Silenced)(nil), errors.New("hi")).
					Once()
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only start the HTTP server here to prevent data races in tests
			server := httptest.NewServer(parentRouter)
			defer server.Close()

			if tt.controllerFunc != nil {
				tt.controllerFunc(controller)
			}

			// Prepare the HTTP request
			client := new(http.Client)
			req, err := http.NewRequest(tt.method, server.URL+tt.path, bytes.NewReader(tt.body))
			if err != nil {
				t.Fatal(err)
			}

			// Perform the HTTP request
			res, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()

			// Inspect the response code
			if res.StatusCode != tt.wantStatusCode {
				t.Errorf("EventsRouter StatusCode = %v, wantStatusCode %v", res.StatusCode, tt.wantStatusCode)
				body, _ := ioutil.ReadAll(res.Body)
				t.Errorf("error message: %q", string(body))
				return
			}
		})
	}
}
