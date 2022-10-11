package resource

import (
	"context"
	"errors"
	"testing"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEntityStore struct {
	mock.Mock
}

func (m *mockEntityStore) DeleteEntity(_ context.Context, _ *corev2.Entity) error {
	panic("should not be called")
}

func (m *mockEntityStore) DeleteEntityByName(_ context.Context, _ string) error {
	panic("should not be called")
}

func (m *mockEntityStore) GetEntities(_ context.Context, _ *store.SelectionPredicate) ([]*corev2.Entity, error) {
	panic("should not be called")
}

func (m *mockEntityStore) GetEntityByName(_ context.Context, _ string) (*corev2.Entity, error) {
	panic("should not be called")
}

func (m *mockEntityStore) UpdateEntity(ctx context.Context, entity *corev2.Entity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

type mockNamespaceStore struct {
	mock.Mock
}

func (m *mockNamespaceStore) CreateNamespace(ctx context.Context, namespace *corev2.Namespace) error {
	args := m.Called(ctx, namespace)
	return args.Error(0)
}

func (m *mockNamespaceStore) DeleteNamespace(_ context.Context, _ string) error {
	panic("should not be called")
}

func (m *mockNamespaceStore) ListNamespaces(_ context.Context, _ *store.SelectionPredicate) ([]*corev2.Namespace, error) {
	panic("should not be called")
}

func (m *mockNamespaceStore) GetNamespace(ctx context.Context, name string) (*corev2.Namespace, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*corev2.Namespace), args.Error(1)
}

func (m *mockNamespaceStore) UpdateNamespace(_ context.Context, _ *corev2.Namespace) error {
	panic("should not be called")
}

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
	mes := &mockEntityStore{}
	mns := &mockNamespaceStore{}
	mmb := &mockMessageBus{published: []*corev2.Event{}}

	br := New(mns, mes, mmb)
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
