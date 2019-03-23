package routers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
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

func (m *mockUserController) Query(ctx context.Context) ([]*types.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*types.User), args.Error(1)
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

func newUserTest(t *testing.T) (*mockUserController, *httptest.Server) {
	ctrl := &mockUserController{}
	routes := &UsersRouter{controller: ctrl}
	router := mux.NewRouter()
	routes.Mount(router)

	return ctrl, httptest.NewServer(router)
}

func TestGetUser(t *testing.T) {
	ctrl, server := newUserTest(t)
	defer server.Close()

	endpoint := "/users"
	client := &http.Client{}

	// User found
	user := types.FixtureUser("A")
	ctrl.On("Find", mock.Anything, "A").Return(user, nil).Once()
	req := newRequest(t, http.MethodGet, server.URL+path.Join(endpoint, "A"), nil)
	res, err := client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	// Not found
	ctrl.On("Find", mock.Anything, "B").Return(nil, actions.NewErrorf(actions.NotFound)).Once()
	req = newRequest(t, http.MethodGet, server.URL+path.Join(endpoint, "B"), nil)
	res, err = client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, 404, res.StatusCode)
}
