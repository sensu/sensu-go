package graphqlschema

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type EntityIDResolver struct {
	suite.Suite
	resolverSuite
}

func (t *EntityIDResolver) TestStandard() {
	record := types.FixtureEntity("name")
	res, err := t.runResolver("Entity.id", record)
	t.NotEmpty(res)
	t.NoError(err)
}

type EntityUserResolver struct {
	suite.Suite
	resolverSuite
}

func (t *EntityUserResolver) TestStandard() {
	user := types.FixtureUser("username")
	record := types.FixtureEntity("name")
	record.User = user.Username
	t.store().On("GetUser", user.Username).Return(user, nil).Once()

	res, err := t.runResolver("Entity.user", record)
	t.NotEmpty(res)
	t.NoError(err)
}

func (t *EntityUserResolver) TestStoreError() {
	user := types.FixtureUser("username")
	record := types.FixtureEntity("name")
	record.User = user.Username
	t.store().On("GetUser", user.Username).Return(user, errMock).Once()

	res, err := t.runResolver("Entity.user", record)
	t.Nil(res)
	t.Error(err)
}

type EntityNodeResolver struct {
	suite.Suite
	nodeSuite
}

func (t *EntityNodeResolver) SetupTest() {
	nodeResolver := newEntityNodeResolver()
	t.setNodeResolver(&nodeResolver)
}

func (t *EntityNodeResolver) TestNoErrors() {
	record := types.FixtureEntity("bob")
	t.store().
		On("GetEntityByID", mock.Anything, record.ID).
		Return(record, nil).Once()

	res, err := t.runResolver(record)
	t.Equal(record, res)
	t.NoError(err)
}

func (t *EntityNodeResolver) TestWithUnauthorizedUser() {
	record := types.FixtureEntity("bob")
	params := t.newParams(record, contextWithNoAccess)
	t.store().
		On("GetEntityByID", mock.Anything, record.ID).
		Return(record, nil).Once()

	res, err := t.runResolver(params)
	t.Nil(res)
	t.NoError(err)
}

func (t *EntityNodeResolver) TestWithStoreError() {
	record := types.FixtureEntity("bob")
	t.store().
		On("GetEntityByID", mock.Anything, record.ID).
		Return(record, errMock).Once()

	res, err := t.runResolver(record)
	t.Nil(res)
	t.Error(err)
}

func TestEntityType(testRunner *testing.T) {
	runSuites(
		testRunner,
		new(EntityIDResolver),
		new(EntityUserResolver),
		new(EntityNodeResolver),
	)
}
