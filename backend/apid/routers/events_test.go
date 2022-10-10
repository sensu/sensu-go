package routers

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/mock"
)

type mockEventController struct {
	mock.Mock
}

func (m *mockEventController) CreateOrReplace(ctx context.Context, check *corev2.Event) error {
	return m.Called(ctx, check).Error(0)
}

func (m *mockEventController) Delete(ctx context.Context, entity, check string) error {
	return m.Called(ctx, entity, check).Error(0)
}

func (m *mockEventController) Get(ctx context.Context, entity, check string) (*corev2.Event, error) {
	args := m.Called(ctx, entity, check)
	return args.Get(0).(*corev2.Event), args.Error(1)
}

func (m *mockEventController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	args := m.Called(ctx, pred)
	return args.Get(0).([]corev2.Resource), args.Error(1)
}

func TestEventsRouter(t *testing.T) {
	type controllerFunc func(*mockEventController)

	// Setup the router
	controller := &mockEventController{}
	router := EventsRouter{controller: controller}
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
	router.Mount(parentRouter)

	empty := &corev2.Event{
		Check:  &corev2.Check{ObjectMeta: corev2.ObjectMeta{}},
		Entity: &corev2.Entity{ObjectMeta: corev2.ObjectMeta{Namespace: "default"}},
	}
	fixture := corev2.FixtureEvent("foo", "check-cpu")

	fixture_wo_namespace := corev2.FixtureEvent("foo", "check-cpu")
	fixture_wo_namespace.Entity.ObjectMeta.Namespace = ""

	tests := []struct {
		name           string
		method         string
		path           string
		body           []byte
		controllerFunc controllerFunc
		wantStatusCode int
	}{
		//
		// GET
		//
		{
			name:   "it returns 404 if a resource is not found",
			method: http.MethodGet,
			path:   fixture.URIPath(),
			controllerFunc: func(c *mockEventController) {
				c.On("Get", mock.Anything, "foo", "check-cpu").
					Return(empty, actions.NewErrorf(actions.NotFound)).
					Once()
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:   "it returns 200 if a resource is found",
			method: http.MethodGet,
			path:   fixture.URIPath(),
			controllerFunc: func(c *mockEventController) {
				c.On("Get", mock.Anything, "foo", "check-cpu").
					Return(fixture, nil).
					Once()
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "it returns 500 if the store encounters an error while listing events",
			method: http.MethodGet,
			path:   empty.URIPath(),
			controllerFunc: func(c *mockEventController) {
				c.On("List", mock.Anything, mock.AnythingOfType("*store.SelectionPredicate")).
					Return([]corev2.Resource{empty}, actions.NewErrorf(actions.InternalErr)).
					Once()
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "it returns 200 and lists resources",
			method: http.MethodGet,
			path:   empty.URIPath(),
			controllerFunc: func(c *mockEventController) {
				c.On("List", mock.Anything, mock.AnythingOfType("*store.SelectionPredicate")).
					Return([]corev2.Resource{fixture}, nil).
					Once()
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "it returns 400 if the payload to create is not decodable",
			method:         http.MethodPost,
			path:           empty.URIPath(),
			body:           []byte(`foo`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "it returns 400 if the event to create is not valid",
			method: http.MethodPost,
			path:   empty.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockEventController) {
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
					Return(actions.NewErrorf(actions.InvalidArgument)).
					Once()
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "it returns 500 if the store returns an error while creating an event",
			method: http.MethodPost,
			path:   empty.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockEventController) {
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
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
			controllerFunc: func(c *mockEventController) {
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
			wantStatusCode: http.StatusCreated,
		},
		//
		// PUT
		//
		{
			name:           "it returns 400 if the payload to update is not decodable",
			method:         http.MethodPut,
			path:           fixture.URIPath(),
			body:           []byte(`foo`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "it returns 400 if the event metadata to update is invalid",
			method:         http.MethodPut,
			path:           fixture.URIPath(),
			body:           []byte(`{"entity": {"metadata": {"namespace":"acme"}}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "it returns 400 if the event metadata to update is invalid",
			method:         http.MethodPut,
			path:           fixture.URIPath(),
			body:           []byte(`{"entity": {}, "check": {"metadata": {"namespace":"acme"}}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "it returns 400 if the event to update is not valid",
			method: http.MethodPut,
			path:   fixture.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockEventController) {
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
					Return(actions.NewErrorf(actions.InvalidArgument)).
					Once()
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "it returns 500 if the store returns an error while updating an event",
			method: http.MethodPut,
			path:   fixture.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockEventController) {
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
					Return(actions.NewErrorf(actions.InternalErr)).
					Once()
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "it returns 201 when an event is successfully updated",
			method: http.MethodPut,
			path:   fixture.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockEventController) {
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:   "it returns 201 when an event does not provide a namespace (using url namespace)",
			method: http.MethodPut,
			path:   fixture.URIPath(),
			body:   marshal(fixture_wo_namespace),
			controllerFunc: func(c *mockEventController) {
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "it returns 400 if the entity name to PUT does not match the URL parameter",
			method:         http.MethodPut,
			path:           fixture.URIPath(),
			body:           []byte(`{"entity": {"metadata": {"name": "bar", "namespace": "default"}}, "check": {"metadata": {"name": "check-cpu", "namespace":"default"}}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "it returns 400 if the entity namespace to PUT does not match the URL parameter",
			method:         http.MethodPut,
			path:           fixture.URIPath(),
			body:           []byte(`{"entity": {"metadata": {"name": "foo", "namespace": "dev"}}, "check": {"metadata": {"name": "check-cpu", "namespace":"default"}}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "it returns 400 if the check name to PUT does not match the URL parameter",
			method:         http.MethodPut,
			path:           fixture.URIPath(),
			body:           []byte(`{"entity": {"metadata": {"name": "foo", "namespace": "default"}}, "check": {"metadata": {"name": "check-mem", "namespace":"default"}}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "it returns 400 if the check namespace to PUT does not match the URL parameter",
			method:         http.MethodPut,
			path:           fixture.URIPath(),
			body:           []byte(`{"entity": {"metadata": {"name": "foo", "namespace": "default"}}, "check": {"metadata": {"name": "check-cpu", "namespace":"dev"}}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		//
		// POST
		//
		{
			name:           "it returns 400 if the payload to update is not decodable (post)",
			method:         http.MethodPost,
			path:           fixture.URIPath(),
			body:           []byte(`foo`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "it returns 400 if the event metadata to update is invalid (post)",
			method:         http.MethodPost,
			path:           fixture.URIPath(),
			body:           []byte(`{"entity": {"metadata": {"namespace":"acme"}}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "it returns 400 if the event metadata to POST is invalid (post)",
			method:         http.MethodPost,
			path:           fixture.URIPath(),
			body:           []byte(`{"entity": {}, "check": {"metadata": {"namespace":"acme"}}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "it returns 400 if the event to update is invalid (post)",
			method: http.MethodPost,
			path:   fixture.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockEventController) {
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
					Return(actions.NewErrorf(actions.InvalidArgument)).
					Once()
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "it returns 500 if the store returns an error while updating an event (post)",
			method: http.MethodPost,
			path:   fixture.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockEventController) {
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
					Return(actions.NewErrorf(actions.InternalErr)).
					Once()
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "it returns 201 when an event is successfully updated (post)",
			method: http.MethodPost,
			path:   fixture.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockEventController) {
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "it returns 400 if the entity name to POST does not match the URL parameter",
			method:         http.MethodPost,
			path:           fixture.URIPath(),
			body:           []byte(`{"entity": {"metadata": {"name": "bar", "namespace": "default"}}, "check": {"metadata": {"name": "check-cpu", "namespace":"default"}}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "it returns 400 if the entity namespace to POST does not match the URL parameter",
			method:         http.MethodPost,
			path:           fixture.URIPath(),
			body:           []byte(`{"entity": {"metadata": {"name": "foo", "namespace": "dev"}}, "check": {"metadata": {"name": "check-cpu", "namespace":"default"}}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "it returns 400 if the check name to POST does not match the URL parameter",
			method:         http.MethodPost,
			path:           fixture.URIPath(),
			body:           []byte(`{"entity": {"metadata": {"name": "foo", "namespace": "default"}}, "check": {"metadata": {"name": "check-mem", "namespace":"default"}}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "it returns 400 if the check namespace to POST does not match the URL parameter",
			method:         http.MethodPost,
			path:           fixture.URIPath(),
			body:           []byte(`{"entity": {"metadata": {"name": "foo", "namespace": "default"}}, "check": {"metadata": {"name": "check-cpu", "namespace":"dev"}}}`),
			wantStatusCode: http.StatusBadRequest,
		},
		//
		// DELETE
		//
		{
			name:   "it returns 404 if the event to delete does not exist",
			method: http.MethodDelete,
			path:   fixture.URIPath(),
			controllerFunc: func(c *mockEventController) {
				c.On("Delete", mock.Anything, "foo", "check-cpu").
					Return(actions.NewErrorf(actions.NotFound)).
					Once()
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:   "it returns 500 if the store returns an error while deleting an event",
			method: http.MethodDelete,
			path:   fixture.URIPath(),
			controllerFunc: func(c *mockEventController) {
				c.On("Delete", mock.Anything, "foo", "check-cpu").
					Return(actions.NewErrorf(actions.InternalErr)).
					Once()
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "it returns 204 if the event was deleted",
			method: http.MethodDelete,
			path:   fixture.URIPath(),
			controllerFunc: func(c *mockEventController) {
				c.On("Delete", mock.Anything, "foo", "check-cpu").
					Return(nil).
					Once()
			},
			wantStatusCode: http.StatusNoContent,
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
