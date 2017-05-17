package basic

import (
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUser(t *testing.T) {
	b := &Basic{}

	store := &mockstore.MockStore{}
	b.Store = store

	store.On("CreateUser", mock.AnythingOfType("*types.User")).Return(nil)

	user := types.FixtureUser("foo")
	user.Password = "P@ssw0rd!"
	assert.NoError(t, b.CreateUser(user))
}

func TestHashPassword(t *testing.T) {
	password := "P@ssw0rd!"

	hash, err := hashPassword(password)
	assert.NotEqual(t, password, hash)
	assert.NoError(t, err)
}
