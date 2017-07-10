package backend

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/testing/mockprovider"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSeedDefaultRole(t *testing.T) {
	store := &mockstore.MockStore{}
	store.On("UpdateRole", mock.AnythingOfType("*types.Role")).Return(nil)
	store.On(
		"UpdateOrganization",
		mock.Anything,
		mock.AnythingOfType("*types.Organization"),
	).Return(nil)

	authProvider := &mockprovider.MockProvider{}
	authProvider.On("CreateUser", mock.AnythingOfType("*types.User")).Return(nil)

	seedInitialData(store, authProvider)
	store.AssertCalled(t, "UpdateRole", mock.AnythingOfType("*types.Role"))
}

func TestSeedDefaultRoleWithError(t *testing.T) {
	assert := assert.New(t)
	store := &mockstore.MockStore{}
	authProvider := &mockprovider.MockProvider{}

	store.On("UpdateRole", mock.AnythingOfType("*types.Role")).Return(errors.New(""))
	assert.Error(seedInitialData(store, authProvider))
}
