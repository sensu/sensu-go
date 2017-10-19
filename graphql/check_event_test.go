package graphqlschema

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type CheckEventIDResolver struct {
	suite.Suite
	resolverSuite
}

func (t *CheckEventIDResolver) TestStandard() {
	event := types.FixtureEvent("entity", "name")
	res, err := t.runResolver("CheckEvent.id", event)
	t.NotEmpty(res)
	t.NoError(err)
}

type CheckEventEntityResolver struct {
	suite.Suite
	resolverSuite
}

func (t *CheckEventEntityResolver) TestStandard() {
	event := types.FixtureEvent("entity", "name")
	res, err := t.runResolver("CheckEvent.entity", event)
	t.NotEmpty(res)
	t.NoError(err)
}

func (t *CheckEventEntityResolver) TestWithUnauthorizedUser() {
	event := types.FixtureEvent("entity", "name")
	params := t.newParams(event, contextWithNoAccess)
	res, err := t.runResolver("CheckEvent.entity", params)
	t.Empty(res)
	t.NoError(err)
}

type CheckEventConfigResolver struct {
	suite.Suite
	resolverSuite
}

func (t *CheckEventConfigResolver) TestStandard() {
	event := types.FixtureEvent("entity", "name")
	res, err := t.runResolver("CheckEvent.config", event)
	t.NotEmpty(res)
	t.NoError(err)
}

func (t *CheckEventConfigResolver) TestWithUnauthorizedUser() {
	event := types.FixtureEvent("entity", "name")
	params := t.newParams(event, contextWithNoAccess)
	res, err := t.runResolver("CheckEvent.config", params)
	t.Empty(res)
	t.NoError(err)
}

type CheckEventNodeResolver struct {
	suite.Suite
	nodeSuite
}

func (t *CheckEventNodeResolver) SetupTest() {
	nodeResolver := newCheckEventNodeResolver()
	t.setNodeResolver(&nodeResolver)
}

func (t *CheckEventNodeResolver) TestNoErrors() {
	record := types.FixtureEvent("one", "two")
	t.store().
		On("GetEventsByEntity", mock.Anything, record.Entity.ID).
		Return([]*types.Event{record}, nil).Once()

	res, err := t.runResolver(record)
	t.Equal(record, res)
	t.NoError(err)
}

func (t *CheckEventNodeResolver) TestWithStoreError() {
	record := types.FixtureEvent("one", "two")
	t.store().
		On("GetEventsByEntity", mock.Anything, record.Entity.ID).
		Return([]*types.Event{record}, errMock).Once()

	res, err := t.runResolver(record)
	t.Empty(res)
	t.Error(err)
}

func (t *CheckEventNodeResolver) TestWithUnauthorizedUser() {
	record := types.FixtureEvent("one", "two")
	t.store().
		On("GetEventsByEntity", mock.Anything, record.Entity.ID).
		Return([]*types.Event{record}, nil).Once()

	params := t.newParams(record, contextWithNoAccess)
	res, err := t.runResolver(params)
	t.Empty(res)
	t.NoError(err)
}

func TestCheckEventType(testRunner *testing.T) {
	runSuites(
		testRunner,
		new(CheckEventNodeResolver),
		new(CheckEventIDResolver),
		new(CheckEventEntityResolver),
		new(CheckEventConfigResolver),
	)
}
