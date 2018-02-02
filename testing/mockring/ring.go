package mockring

import (
	"context"
	"path"

	"github.com/sensu/sensu-go/backend/ring"
	"github.com/stretchr/testify/mock"
)

// Getter ...
type Getter map[string]ring.Interface

// GetRing ...
func (g Getter) GetRing(p ...string) ring.Interface {
	s := path.Join(p...)
	return g[s]
}

// Ring ...
type Ring struct {
	mock.Mock
}

// Add ...
func (r *Ring) Add(ctx context.Context, value string) error {
	args := r.Called(ctx, value)
	return args.Error(0)
}

// Remove ...
func (r *Ring) Remove(ctx context.Context, value string) error {
	args := r.Called(ctx, value)
	return args.Error(0)
}

// Next ...
func (r *Ring) Next(ctx context.Context) (string, error) {
	args := r.Called(ctx)
	return args.Get(0).(string), args.Error(1)
}

// Peek ...
func (r *Ring) Peek(ctx context.Context) (string, error) {
	args := r.Called(ctx)
	return args.Get(0).(string), args.Error(1)
}
