package etcd

import (
	"context"
	"errors"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

const (
	pipelinesPathPrefix = "pipelines"
)

var (
	pipelineKeyBuilder = store.NewKeyBuilder(pipelinesPathPrefix)
)

// TODO(jk): remove the nolint line after the pipeline store is complete
//nolint:deadcode,unused
func getPipelinePath(pipeline *corev2.Pipeline) string {
	return pipelineKeyBuilder.WithResource(pipeline).Build(pipeline.Name)
}

// GetPipelinesPath gets the path of the check config store.
func GetPipelinesPath(ctx context.Context, name string) string {
	return pipelineKeyBuilder.WithContext(ctx).Build(name)
}

// GetPipelineByName gets a Pipeline by name.
func (s *Store) GetPipelineByName(ctx context.Context, name string) (*corev2.Pipeline, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	var pipeline corev2.Pipeline
	if err := Get(ctx, s.client, GetPipelinesPath(ctx, name), &pipeline); err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
		return nil, err
	}

	return &pipeline, nil
}
