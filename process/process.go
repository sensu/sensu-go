package process

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// A Getter is responsible for getting the process info of an agent.
type Getter interface {
	Get(context.Context) ([]*corev2.Process, error)
}

// A NoopProcessGetter is responsible for refreshing process info of an agent.
type NoopProcessGetter struct{}

// Get is not yet implemented.
func (n *NoopProcessGetter) Get(ctx context.Context) ([]*corev2.Process, error) {
	return []*corev2.Process{}, nil
}
