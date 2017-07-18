package keepalived

import (
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type KeepalivedTestSuite struct {
	suite.Suite
	Keepalived   *Keepalived
	MessageBus   messaging.MessageBus
	Store        *mockstore.MockStore
	Deregisterer *mockDeregisterer
	EventCreator *mockCreator
}

func (suite *KeepalivedTestSuite) SetupTest() {
	suite.MessageBus = &messaging.WizardBus{}
	suite.MessageBus.Start()

	mockStore := &mockstore.MockStore{}
	dereg := &mockDeregisterer{}
	creator := &mockCreator{}

	suite.Deregisterer = dereg
	suite.EventCreator = creator
	suite.Store = mockStore

	keepalived := &Keepalived{
		Store:      suite.Store,
		MessageBus: suite.MessageBus,
		MonitorFactory: func(e *types.Entity) *KeepaliveMonitor {
			return &KeepaliveMonitor{
				Entity:       e,
				Deregisterer: dereg,
				EventCreator: creator,
				Store:        mockStore,
			}
		},
	}

	suite.Keepalived = keepalived
}

func (suite *KeepalivedTestSuite) AfterTest() {
	suite.MessageBus.Stop()
	suite.Keepalived.Stop()
}

func (suite *KeepalivedTestSuite) TestStartStop() {
	k := &Keepalived{}
	suite.Error(k.Start())

	k.MessageBus = suite.MessageBus
	suite.Error(k.Start())

	k.MonitorFactory = nil

	store := &mockstore.MockStore{}
	store.On("GetFailingKeepalives", mock.Anything).Return([]*types.KeepaliveRecord{}, nil)

	k.Store = store
	suite.NoError(k.Start())
	suite.NotNil(k.MonitorFactory, "*Keepalived.Start() ensures there is a MonitorFactory")

	suite.NoError(k.Status())

	var err error
	select {
	case err = <-k.Err():
	default:
	}
	suite.NoError(err)

	suite.NoError(k.Stop())
}

func (suite *KeepalivedTestSuite) TestEventProcessing() {
	suite.Store.On("GetFailingKeepalives", mock.Anything).Return([]*types.KeepaliveRecord{}, nil)
	suite.Keepalived.MonitorFactory = nil
	suite.NoError(suite.Keepalived.Start())
	event := types.FixtureEvent("check", "entity")
	suite.Store.On("UpdateEntity", mock.Anything, event.Entity).Return(nil)
	suite.Keepalived.keepaliveChan <- event
	time.Sleep(100 * time.Millisecond)
	suite.Store.AssertCalled(suite.T(), "UpdateEntity", mock.Anything, event.Entity)
}

func TestKeepalivedSuite(t *testing.T) {
	suite.Run(t, new(KeepalivedTestSuite))
}
