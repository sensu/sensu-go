package ringv2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockRing struct {
	mock.Mock
}

func (m *MockRing) Subscribe(ctx context.Context, sub Subscription) <-chan Event {
	return m.Called(ctx, sub).Get(0).(<-chan Event)
}

func (m *MockRing) Remove(ctx context.Context, value string) error {
	return m.Called(ctx, value).Error(0)
}

func (m *MockRing) Add(ctx context.Context, value string, keepalive int64) error {
	return m.Called(ctx, value, keepalive).Error(0)
}

func (m *MockRing) IsEmpty(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

func TestPoolSetNewFunc(t *testing.T) {
	pool := NewRingPool(func(path string) Interface {
		return new(MockRing)
	})
	fooRing := pool.Get("foo")

	pool.SetNewFunc(func(path string) Interface {
		return nil
	})

	fooRing2 := pool.Get("foo")

	if fooRing == fooRing2 {
		t.Fatal("rings should differ")
	}
}
