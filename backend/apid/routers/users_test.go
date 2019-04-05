package routers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserController struct {
	mock.Mock
}

func (m *mockUserController) Create(ctx context.Context, name types.User) error {
	return m.Called(ctx, name).Error(0)
}

func (m *mockUserController) CreateOrReplace(ctx context.Context, name types.User) error {
	return m.Called(ctx, name).Error(0)
}

func (m *mockUserController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	args := m.Called(ctx, pred)
	return args.Get(0).([]corev2.Resource), args.Error(1)
}

func (m *mockUserController) Find(ctx context.Context, name string) (*types.User, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.User), args.Error(1)
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

func newUserTest() (*mockUserController, *httptest.Server) {
	ctrl := &mockUserController{}
	routes := &UsersRouter{controller: ctrl}
	router := mux.NewRouter()
	routes.Mount(router)

	return ctrl, httptest.NewServer(router)
}

func TestGetUser(t *testing.T) {
	fixture := types.FixtureUser("fred")
	endpoint := "/users"
	client := &http.Client{}

	testCases := []struct {
		path       string
		setup      func(ctrl *mockUserController)
		statusCode int
	}{
		{
			path: path.Join(endpoint, fixture.Username),
			setup: func(ctrl *mockUserController) {
				ctrl.On("Find", mock.Anything, mock.Anything).Return(fixture, nil)
			},
			statusCode: 200,
		},
		{
			path: path.Join(endpoint, "bob"),
			setup: func(ctrl *mockUserController) {
				err := actions.NewErrorf(actions.NotFound)
				ctrl.On("Find", mock.Anything, mock.Anything).Return(nil, err)
			},
			statusCode: 404,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			ctrl, server := newUserTest()
			defer server.Close()

			tc.setup(ctrl)

			req := newRequest(t, http.MethodGet, server.URL+tc.path, nil)
			res, err := client.Do(req)
			require.NoError(t, err)
			assert.Equal(t, tc.statusCode, res.StatusCode)
		})
	}
}
