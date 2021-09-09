package mockpipeline

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/mock"
)

// FilterAdapter ...
type FilterAdapter struct {
	mock.Mock
}

// Name ...
func (m FilterAdapter) Name() string {
	args := m.Called()
	return args.Get(0).(string)
}

// CanMutate ...
func (m FilterAdapter) CanFilter(ref *corev2.ResourceReference) bool {
	args := m.Called(ref)
	return args.Get(0).(bool)
}

// Filter ...
func (m FilterAdapter) Filter(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (bool, error) {
	args := m.Called(ctx, ref, event)
	return args.Get(0).(bool), args.Error(1)
}
