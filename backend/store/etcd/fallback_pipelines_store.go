package etcd

import (
	"context"
	"errors"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// GetFallbackPipelines retrieves fallback pipelines by name.
func (s *Store) GetFallbackPipelines(ctx context.Context, name string) ([]*corev2.Pipeline, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name of fallback pipelines")}
	}

	// TODO - implement the logic to retrieve all the pipeline from store and return the slice of pipelines

	var f corev2.Pipeline
	if err := Get(ctx, s.client, GetPipelinesPath(ctx, name), &f); err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
		return nil, err
	}

	return []*corev2.Pipeline{&f}, nil
}
