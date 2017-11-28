package graphqlschema

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type HookIDResolver struct {
	suite.Suite
	resolverSuite
}

func (t *HookIDResolver) TestStandard() {
	hook := types.FixtureHookConfig("name")
	res, err := t.runResolver("HookConfig.id", hook)
	t.NotEmpty(res)
	t.NoError(err)
}

type HookNodeResolver struct {
	suite.Suite
	nodeSuite
}

func (t *HookNodeResolver) SetupTest() {
	nodeResolver := newHookConfigNodeResolver()
	t.setNodeResolver(&nodeResolver)
}

func (t *HookNodeResolver) TestNoErrors() {
	record := types.FixtureHookConfig("bob")
	t.store().
		On("GetHookConfigByName", mock.Anything, record.Name).
		Return(record, nil).Once()

	res, err := t.runResolver(record)
	t.Equal(record, res)
	t.NoError(err)
}

func (t *HookNodeResolver) TestWithUnauthorizedUser() {
	record := types.FixtureHookConfig("bob")
	params := t.newParams(record, contextWithNoAccess)
	t.store().
		On("GetHookConfigByName", mock.Anything, record.Name).
		Return(record, nil).Once()

	res, err := t.runResolver(params)
	t.Nil(res)
	t.NoError(err)
}

func (t *HookNodeResolver) TestWithStoreError() {
	record := types.FixtureHookConfig("bob")
	t.store().
		On("GetHookConfigByName", mock.Anything, record.Name).
		Return(record, errMock).Once()

	res, err := t.runResolver(record)
	t.Nil(res)
	t.Error(err)
}

func TestHookConfigType(testRunner *testing.T) {
	runSuites(
		testRunner,
		new(HookIDResolver),
		new(HookNodeResolver),
	)
}
