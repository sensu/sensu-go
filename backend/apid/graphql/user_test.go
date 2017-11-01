package graphqlschema

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/suite"
)

type UserTypeIDFieldSuite struct {
	suite.Suite
	resolverSuite
}

func (t *UserTypeIDFieldSuite) TestEncode() {
	user := types.FixtureUser("bob")
	id, err := t.runResolver("User.id", user)

	t.NotEmpty(id)
	t.NoError(err)
}

type UserNodeSuite struct {
	suite.Suite
	nodeSuite
}

func (t *UserNodeSuite) SetupTest() {
	nodeResolver := newUserNodeResolver()
	t.setNodeResolver(&nodeResolver)
}

func (t *UserNodeSuite) TestNoErrors() {
	record := types.FixtureUser("bob")
	t.store().On("GetUser", record.Username).Return(record, nil).Once()

	res, err := t.runResolver(record)
	t.Equal(record, res)
	t.NoError(err)
}

func (t *UserNodeSuite) TestWithUnauthorizedUser() {
	record := types.FixtureUser("bob")
	params := t.newParams(record, contextWithNoAccess)
	t.store().On("GetUser", record.Username).Return(record, nil).Once()

	res, err := t.runResolver(params)
	t.Nil(res)
	t.NoError(err)
}

func (t *UserNodeSuite) TestWithStoreError() {
	record := types.FixtureUser("bob")
	t.store().
		On("GetUser", record.Username).
		Return(record, errors.New("poopy")).
		Once()

	res, err := t.runResolver(record)
	t.Nil(res)
	t.Error(err)
}

func TestUserType(testRunner *testing.T) {
	runSuites(
		testRunner,
		new(UserTypeIDFieldSuite),
		new(UserNodeSuite),
	)
}
