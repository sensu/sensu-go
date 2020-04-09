package process

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// A Getter is responsible for getting the process info of an agent.
type Getter interface {
	Get(context.Context) ([]*corev2.Process, error)
}

// A ProcGetter is responsible for refreshing process info of an agent.
type ProcGetter struct{}

// Get is not yet implemented.
func (p *ProcGetter) Get(ctx context.Context) ([]*corev2.Process, error) {
	return nil, nil
}
