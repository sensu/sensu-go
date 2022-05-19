package poll

import (
	"context"
	"testing"
	"time"

	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPolling(t *testing.T) {

	t.Run("watcher closes channel on context cancel", func(t *testing.T) {
		pollerUnderTest := Poller{
			Interval:  time.Millisecond * 10,
			TxnWindow: time.Millisecond * 10,
		}
		inactiveTable := &stubTable{}
		t0 := time.Now()
		inactiveTable.On("Now", mock.Anything).Return(t0, nil).Once()
		inactiveTable.On("Since", mock.Anything, t0).Return(make([]Row, 0), nil)
		pollerUnderTest.Table = inactiveTable

		events := make(chan storev2.WatchEvent, 16)
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer cancel()
		go pollerUnderTest.Watch(ctx, events)
		select {
		case e, ok := <-events:
			if !ok {
				return
			}
			t.Errorf("unexpected event: %v", e)
		case <-time.After(350 * time.Millisecond):
			t.Fatalf("expected polling watcher to close channel on context deadline")
		}
	})

	t.Run("watcher publishes updates", func(t *testing.T) {
		pollerUnderTest := Poller{
			Interval:  time.Millisecond * 10,
			TxnWindow: time.Millisecond * 10,
		}
		singleUpdateTable := &stubTable{}
		t0 := time.Now()
		t1 := t0.Add(time.Millisecond * 3)
		singleUpdateTable.On("Now", mock.Anything).Return(t0, nil).Once()
		singleUpdateTable.On("Since", mock.Anything, t0).Return([]Row{forgeRow(t1)}, nil)
		pollerUnderTest.Table = singleUpdateTable

		events := make(chan storev2.WatchEvent, 16)
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
		defer cancel()
		go pollerUnderTest.Watch(ctx, events)
		e := <-events
		assert.Equal(t, storev2.Update, e.Action)
		select {
		case e, ok := <-events:
			if !ok {
				return
			}
			t.Errorf("unexpected event: %v", e)
		case <-time.After(400 * time.Millisecond):
			t.Fatalf("expected polling watcher to close channel on context deadline")
		}
	})
}

type stubTable struct {
	mock.Mock
}

func (s *stubTable) Now(ctx context.Context) (time.Time, error) {
	args := s.Called(ctx)
	return args.Get(0).(time.Time), args.Error(1)
}

func (s *stubTable) Since(ctx context.Context, ts time.Time) ([]Row, error) {
	args := s.Called(ctx, ts)
	return args.Get(0).([]Row), args.Error(1)
}

func forgeRow(ts time.Time) Row {
	return Row{
		CreatedAt: ts.Add(-1 * time.Second),
		UpdatedAt: ts,
		Id:        "fake",
	}
}
