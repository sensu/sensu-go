package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/mock"
)

type mockListener struct {
	mock.Mock
}

func (m *mockListener) Listen(channel string) error {
	return m.Called(channel).Error(0)
}

func (m *mockListener) Unlisten(channel string) error {
	return m.Called(channel).Error(0)
}

func (m *mockListener) UnlistenAll() error {
	return m.Called().Error(0)
}

func (m *mockListener) Close() error {
	return m.Called().Error(0)
}

func (m *mockListener) NotificationChannel() <-chan *pq.Notification {
	return m.Called().Get(0).(chan *pq.Notification)
}

var _ Listener = new(mockListener)

func TestBusDemux(t *testing.T) {
	listener := new(mockListener)
	listener.On("Listen", ListenChannelName("default", "foo")).Return(nil)
	listener.On("Listen", ListenChannelName("default", "bar")).Return(nil)
	listener.On("Close").Return(nil)
	ch := make(chan *pq.Notification)
	listener.On("NotificationChannel").Return(ch)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bus := NewBus(ctx, listener)
	fooNotes, err := bus.Subscribe(ctx, "default", "foo")
	if err != nil {
		t.Fatal(err)
	}
	barNotes, err := bus.Subscribe(ctx, "default", "bar")
	if err != nil {
		t.Fatal(err)
	}

	ch <- &pq.Notification{
		Channel: ListenChannelName("default", "foo"),
	}

	<-fooNotes

	ch <- &pq.Notification{
		Channel: ListenChannelName("default", "bar"),
	}

	<-barNotes
}

func TestBusDemuxTimeout(t *testing.T) {
	listener := new(mockListener)
	listener.On("Listen", ListenChannelName("default", "abandoned")).Return(nil)
	listener.On("Close").Return(nil)
	ch := make(chan *pq.Notification)
	listener.On("NotificationChannel").Return(ch)

	testCtx, testCancel := context.WithCancel(context.Background())
	defer testCancel()
	bus := NewBus(testCtx, listener)

	// One abandoned subscription
	ctxA, cancelA := context.WithCancel(testCtx)
	defer cancelA()
	_, err := bus.Subscribe(ctxA, "default", "abandoned")
	if err != nil {
		t.Fatal(err)
	}
	// One subscription still being read
	ctx, cancel := context.WithCancel(testCtx)
	defer cancel()
	abandoned, err := bus.Subscribe(ctx, "default", "abandoned")
	if err != nil {
		t.Fatal(err)
	}

	// cancel the abandoned context immediately
	cancelA()
	// allow some time for subscription cancellation to occur
	time.Sleep(5 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		// Send and attempt to read N + 1 notifications to overflow the buffered channel size from newNotificationChan
		for i := 0; i < 9; i++ {
			ch <- &pq.Notification{
				Channel: ListenChannelName("default", "abandoned"),
				Extra:   fmt.Sprintf("notification: %d", i),
			}
			<-abandoned
		}
		done <- struct{}{}
	}()
	select {
	case <-time.After(3 * time.Second):
		t.Fatal("expected to publish after abandoned context canceled")
	case <-done:
		// expected
	}

}
