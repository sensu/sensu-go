package graphqlschema

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CheckIDResolver struct {
	suite.Suite
	resolverSuite
}

func (t *CheckIDResolver) TestStandard() {
	check := types.FixtureCheckConfig("name")
	res, err := t.runResolver("CheckConfig.id", check)
	t.NotEmpty(res)
	t.NoError(err)
}

type CheckNodeResolver struct {
	suite.Suite
	nodeSuite
}

func (t *CheckNodeResolver) SetupTest() {
	nodeResolver := newCheckConfigNodeResolver()
	t.setNodeResolver(&nodeResolver)
}

func (t *CheckNodeResolver) TestNoErrors() {
	record := types.FixtureCheckConfig("bob")
	t.store().
		On("GetCheckConfigByName", mock.Anything, record.Name).
		Return(record, nil).Once()

	res, err := t.runResolver(record)
	t.Equal(record, res)
	t.NoError(err)
}

func (t *CheckNodeResolver) TestWithUnauthorizedUser() {
	record := types.FixtureCheckConfig("bob")
	params := t.newParams(record, contextWithNoAccess)
	t.store().
		On("GetCheckConfigByName", mock.Anything, record.Name).
		Return(record, nil).Once()

	res, err := t.runResolver(params)
	t.Nil(res)
	t.NoError(err)
}

func (t *CheckNodeResolver) TestWithStoreError() {
	record := types.FixtureCheckConfig("bob")
	t.store().
		On("GetCheckConfigByName", mock.Anything, record.Name).
		Return(record, errors.New("poopy")).Once()

	res, err := t.runResolver(record)
	t.Nil(res)
	t.Error(err)
}

func TestCheckConfigType(testRunner *testing.T) {
	runSuites(
		testRunner,
		new(CheckIDResolver),
		new(CheckNodeResolver),
	)
}
