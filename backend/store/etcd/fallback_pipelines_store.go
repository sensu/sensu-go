package etcd

import (
	"context"
	"errors"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

const fallbackPipelinesPathPrefix = "fallback_pipelines"

var (
	fallbackPipelinesKeyBuilder = store.NewKeyBuilder(fallbackPipelinesPathPrefix)
)

// GetFallbackPipelinesPath gets the path of the fallback pipelines store.
func GetFallbackPipelinesPath(ctx context.Context, name string) string {
	return fallbackPipelinesKeyBuilder.WithContext(ctx).Build(name)
}

// GetFallbackPipelines retrieves fallback pipelines by name.
func (s *Store) GetFallbackPipelines(ctx context.Context, name string) ([]*corev2.Pipeline, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name of fallback pipelines")}
	}

	var fp corev2.FallbackPipelines
	if err := Get(ctx, s.client, GetFallbackPipelinesPath(ctx, name), &fp); err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
		return nil, err
	}
	if len(fp.Pipelines) == 0 {
		return []*corev2.Pipeline{}, nil
	}

	var pipelines []*corev2.Pipeline
	var p corev2.Pipeline
	for _, pipeline := range fp.Pipelines {
		if err := Get(ctx, s.client, GetPipelinesPath(ctx, pipeline.Name), &p); err != nil {
			continue
		}
		pipelines = append(pipelines, &p)
	}
	return pipelines, nil
}
