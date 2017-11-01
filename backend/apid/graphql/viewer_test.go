package graphqlschema

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ViewerEntitiesResolver struct {
	suite.Suite
	resolverSuite
}

func (t *ViewerEntitiesResolver) TestNoErrors() {
	t.store().
		On("GetEntities", mock.Anything).
		Return([]*types.Entity{types.FixtureEntity("one")}, nil).
		Once()

	res, err := t.runResolver("Viewer.entities", nil)
	t.NotEmpty(res)
	t.NoError(err)
}

func (t *ViewerEntitiesResolver) TestWithStoreErrors() {
	t.store().
		On("GetEntities", mock.Anything).
		Return([]*types.Entity{}, errors.New("fudge")).
		Once()

	res, err := t.runResolver("Viewer.entities", nil)
	t.Empty(res)
	t.Error(err)
}

type ViewerChecksResolver struct {
	suite.Suite
	resolverSuite
}

func (t *ViewerChecksResolver) TestNoErrors() {
	t.store().
		On("GetCheckConfigs", mock.Anything).
		Return([]*types.CheckConfig{types.FixtureCheckConfig("one")}, nil).
		Once()

	res, err := t.runResolver("Viewer.checks", nil)
	t.NotEmpty(res)
	t.NoError(err)
}

func (t *ViewerChecksResolver) TestWithStoreErrors() {
	t.store().
		On("GetCheckConfigs", mock.Anything).
		Return([]*types.CheckConfig{}, errors.New("oh fudge")).
		Once()

	res, err := t.runResolver("Viewer.checks", nil)
	t.Empty(res)
	t.Error(err)
}

type ViewerCheckEventsResolver struct {
	suite.Suite
	resolverSuite
}

func (t *ViewerCheckEventsResolver) TestNoErrors() {
	t.store().
		On("GetEvents", mock.Anything).
		Return([]*types.Event{types.FixtureEvent("one", "two")}, nil).
		Once()

	res, err := t.runResolver("Viewer.checkEvents", nil)
	t.NotEmpty(res)
	t.NoError(err)
}

func (t *ViewerCheckEventsResolver) TestWithStoreErrors() {
	t.store().
		On("GetEvents", mock.Anything).
		Return([]*types.Event{}, errors.New("oh fudge")).
		Once()

	res, err := t.runResolver("Viewer.checkEvents", nil)
	t.Empty(res)
	t.Error(err)
}

type ViewerUserResolver struct {
	suite.Suite
	resolverSuite
}

func (t *ViewerUserResolver) TestNoErrors() {
	t.store().
		On("GetUser", mock.Anything).
		Return(&types.User{}, nil).
		Once()

	res, err := t.runResolver("Viewer.user", nil)
	t.NotEmpty(res)
	t.NoError(err)
}

func (t *ViewerUserResolver) TestWithStoreErrors() {
	t.store().
		On("GetUser", mock.Anything).
		Return(&types.User{}, errors.New("poop")).
		Once()

	res, err := t.runResolver("Viewer.user", nil)
	t.Empty(res)
	t.Error(err)
}

func TestViewerType(testRunner *testing.T) {
	runSuites(
		testRunner,
		new(ViewerEntitiesResolver),
		new(ViewerChecksResolver),
		new(ViewerCheckEventsResolver),
		new(ViewerUserResolver),
	)
}
