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

	store.On("UpdateUser", mock.AnythingOfType("*types.User")).Return(nil)

	user := types.FixtureUser("foo")
	assert.NoError(t, b.CreateUser(user))
}
