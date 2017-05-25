package basic

import (
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthenticateMissingUser(t *testing.T) {
	b := &Basic{}
	store := &mockstore.MockStore{}
	b.Store = store

	store.On("GetUser", "foo").Return(&types.User{}, fmt.Errorf(""))

	_, err := b.Authenticate("foo", "P@ssw0rd!")
	assert.Error(t, err)
}

func TestAuthenticateDisabledUser(t *testing.T) {
	b := &Basic{}
	store := &mockstore.MockStore{}
	b.Store = store

	user := types.FixtureUser("foo")
	user.Password = "$2a$10$iyYyGmveS9dcYp5DHMbOm.LShX806vB0ClzoPyt1TIgkZ9KQ62cOO"
	user.Disabled = true
	store.On("GetUser", "foo").Return(user, nil)

	_, err := b.Authenticate("foo", "P@ssw0rd!")
	assert.Error(t, err)
}

func TestAuthenticateWrongPassword(t *testing.T) {
	b := &Basic{}
	store := &mockstore.MockStore{}
	b.Store = store

	user := types.FixtureUser("foo")
	user.Password = "$2a$10$iyYyGmveS9dcYp5DHMbOm.LShX806vB0ClzoPyt1TIgkZ9KQ62cOO"
	store.On("GetUser", "foo").Return(user, nil)

	_, err := b.Authenticate("foo", "bar")
	assert.Error(t, err)
}

func TestAuthenticateSuccess(t *testing.T) {
	b := &Basic{}
	store := &mockstore.MockStore{}
	b.Store = store

	user := types.FixtureUser("foo")
	user.Password = "$2a$10$iyYyGmveS9dcYp5DHMbOm.LShX806vB0ClzoPyt1TIgkZ9KQ62cOO"
	store.On("GetUser", "foo").Return(user, nil)

	result, err := b.Authenticate("foo", "P@ssw0rd!")
	assert.NoError(t, err)
	assert.Equal(t, "foo", result.Username)
}

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

func TestCheckPasswod(t *testing.T) {
	hash := "$2a$10$iyYyGmveS9dcYp5DHMbOm.LShX806vB0ClzoPyt1TIgkZ9KQ62cOO"
	password := "P@ssw0rd!"

	assert.False(t, checkPassword(hash, "foo"))
	assert.True(t, checkPassword(hash, password))
}
