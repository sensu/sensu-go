package resource

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store/v2/storetest"
)

type mockMessageBus struct {
	mock.Mock
	published []*corev2.Event
}

func (m *mockMessageBus) Start() error {
	panic("should not be called")
}

func (m *mockMessageBus) Stop() error {
	panic("should not be called")
}

func (m *mockMessageBus) Err() <-chan error {
	panic("should not be called")
}

func (m *mockMessageBus) Name() string {
	panic("should not be called")
}

func (m *mockMessageBus) Subscribe(_ string, _ string, _ messaging.Subscriber) (messaging.Subscription, error) {
	panic("should not be called")
}

func (m *mockMessageBus) Publish(topic string, message interface{}) error {
	args := m.Called(topic, message)
	err := args.Error(0)
	if err == nil {
		m.published = append(m.published, message.(*corev2.Event))
	}
	return err
}

func TestBackendResource_GenerateBackendEvent(t *testing.T) {
	store := &storetest.Store{}
	mmb := &mockMessageBus{published: []*corev2.Event{}}

	br := New(store, mmb)
	br.backendEntity, _ = getEntity()
	br.repeatIntervalSec = 2

	mmb.On("Publish", messaging.TopicEventRaw, mock.MatchedBy(func(input interface{}) bool {
		return input.(*corev2.Event).Check.Output != "6"
	})).Return(nil)
	mmb.On("Publish", messaging.TopicEventRaw, mock.MatchedBy(func(input interface{}) bool {
		return input.(*corev2.Event).Check.Output == "6"
	})).Return(errors.New("oops"))

	// First event - ok status
	err := br.GenerateBackendEvent("secrets", 0, "1")
	assert.NoError(t, err)

	// Repeated immediately - should not publish because of repeat interval
	err = br.GenerateBackendEvent("secrets", 0, "2")
	assert.NoError(t, err)
	time.Sleep(time.Second * 3)

	// Repeated after a pause - should publish
	err = br.GenerateBackendEvent("secrets", 0, "3")
	assert.NoError(t, err)

	// Critical status - should publish immediately
	err = br.GenerateBackendEvent("secrets", 2, "4")
	assert.NoError(t, err)

	// Repeat critical - should not publish
	err = br.GenerateBackendEvent("secrets", 2, "5")
	assert.NoError(t, err)
	time.Sleep(time.Second * 3)

	// Bus error
	err = br.GenerateBackendEvent("secrets", 2, "6")
	assert.Error(t, err)

	assert.Equal(t, 3, len(mmb.published))
	assert.Equal(t, "1", mmb.published[0].Check.Output)
	assert.Equal(t, "3", mmb.published[1].Check.Output)
	assert.Equal(t, "4", mmb.published[2].Check.Output)
}
