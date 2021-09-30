package mockpipeline

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/stretchr/testify/mock"
)

// MutatorAdapter ...
type MutatorAdapter struct {
	mock.Mock
}

// Name ...
func (m *MutatorAdapter) Name() string {
	args := m.Called()
	return args.Get(0).(string)
}

// CanMutate ...
func (m *MutatorAdapter) CanMutate(ref *corev2.ResourceReference) bool {
	args := m.Called(ref)
	return args.Get(0).(bool)
}

// Mutate ...
func (m *MutatorAdapter) Mutate(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) ([]byte, error) {
	args := m.Called(ctx, ref, event)
	return args.Get(0).([]byte), args.Error(1)
}
