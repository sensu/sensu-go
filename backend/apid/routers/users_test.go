package routers

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/mock"
)

type mockUserController struct {
	mock.Mock
}

func (m *mockUserController) AuthenticateUser(ctx context.Context, username, password string) (*corev2.User, error) {
	args := m.Called(ctx, username, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*corev2.User), args.Error(1)
}

func (m *mockUserController) Create(ctx context.Context, user *corev2.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserController) CreateOrReplace(ctx context.Context, user *corev2.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	args := m.Called(ctx, pred)
	return args.Get(0).([]corev2.Resource), args.Error(1)
}

func (m *mockUserController) Get(ctx context.Context, name string) (*corev2.User, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*corev2.User), args.Error(1)
}

func (m *mockUserController) Disable(ctx context.Context, name string) error {
	return m.Called(ctx, name).Error(0)
}

func (m *mockUserController) Enable(ctx context.Context, name string) error {
	return m.Called(ctx, name).Error(0)
}

func (m *mockUserController) AddGroup(ctx context.Context, name string, group string) error {
	return m.Called(ctx, name, group).Error(0)
}

func (m *mockUserController) RemoveGroup(ctx context.Context, name string, group string) error {
	return m.Called(ctx, name, group).Error(0)
}

func (m *mockUserController) RemoveAllGroups(ctx context.Context, name string) error {
	return m.Called(ctx, name).Error(0)
}

func TestUsersRouter(t *testing.T) {
	type controllerFunc func(*mockUserController)

	// Setup the router
	controller := &mockUserController{}
	router := UsersRouter{controller: controller}
	parentRouter := mux.NewRouter().PathPrefix(corev2.URLPrefix).Subrouter()
	router.Mount(parentRouter)

	empty := &corev2.User{}
	fixture := corev2.FixtureUser("foo")

	tests := []struct {
		name           string
		method         string
		path           string
		body           []byte
		controllerFunc controllerFunc
		wantStatusCode int
	}{
		{
			name:   "it returns 404 if a resource is not found",
			method: http.MethodGet,
			path:   fixture.URIPath(),
			controllerFunc: func(c *mockUserController) {
				c.On("Get", mock.Anything, "foo").
					Return(empty, actions.NewErrorf(actions.NotFound)).
					Once()
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:   "it returns 200 if a resource is found",
			method: http.MethodGet,
			path:   fixture.URIPath(),
			controllerFunc: func(c *mockUserController) {
				c.On("Get", mock.Anything, "foo").
					Return(fixture, nil).
					Once()
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:   "it returns 500 if the store encounters an error while listing users",
			method: http.MethodGet,
			path:   empty.URIPath(),
			controllerFunc: func(c *mockUserController) {
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
			controllerFunc: func(c *mockUserController) {
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
			name:   "it returns 400 if the user to create is not valid",
			method: http.MethodPost,
			path:   empty.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockUserController) {
				c.On("Create", mock.Anything, mock.Anything).
					Return(actions.NewErrorf(actions.InvalidArgument)).
					Once()
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "it returns 500 if the store returns an error while creating a user",
			method: http.MethodPost,
			path:   empty.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockUserController) {
				c.On("Create", mock.Anything, mock.Anything).
					Return(actions.NewErrorf(actions.InternalErr)).
					Once()
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:   "it returns 201 when a user is successfully created",
			method: http.MethodPost,
			path:   empty.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockUserController) {
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
			name:           "it returns 400 if the user metadata to update is invalid",
			method:         http.MethodPut,
			path:           fixture.URIPath(),
			body:           []byte(`{"username":"bar"}`),
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:   "it returns 500 if the store returns an error while updating a user",
			method: http.MethodPut,
			path:   fixture.URIPath(),
			body:   marshal(fixture),
			controllerFunc: func(c *mockUserController) {
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
			controllerFunc: func(c *mockUserController) {
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
					Return(nil).
					Once()
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name:   "it only pass the password hash to CreateOrReplace when updating a password",
			method: http.MethodPut,
			path:   path.Join(fixture.URIPath(), "password"),
			body:   []byte(`{"password":"P@ssw0rd!","password_hash":"$2a$10$PdP2LURUHv7PylQtu8haL.8ZBSr5fjDmWXacNGWL6juiR4fRaRSNS"}`),
			controllerFunc: func(c *mockUserController) {
				c.On("AuthenticateUser", mock.Anything, mock.Anything, mock.Anything).
					Return(&corev2.User{Username: "foo", Password: "password_hash", PasswordHash: "password_hash"}, nil).
					Once()
				c.On("CreateOrReplace", mock.Anything, mock.Anything).
					Return(nil).
					Once().
					Run(func(args mock.Arguments) {
						user := args.Get(1).(*corev2.User)
						if user.Password != "" {
							t.Fatal("only the password hash should be passed to the CreateOrReplace controller")
						}
					})
			},
			wantStatusCode: http.StatusCreated,
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
