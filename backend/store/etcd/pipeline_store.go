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

// GetPipelinesPath gets the path of the pipelines store.
func GetPipelinesPath(ctx context.Context, name string) string {
	return pipelineKeyBuilder.WithContext(ctx).Build(name)
}

// DeletePipelineByName deletes a Pipeline by name.
func (s *Store) DeletePipelineByName(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name of pipeline")}
	}

	err := Delete(ctx, s.client, GetPipelinesPath(ctx, name))
	if _, ok := err.(*store.ErrNotFound); ok {
		err = nil
	}

	return err
}

// GetPipelines gets the list of Pipelines for a namespace.
func (s *Store) GetPipelines(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Pipeline, error) {
	pipelines := []*corev2.Pipeline{}
	err := List(ctx, s.client, GetPipelinesPath, &pipelines, pred)
	return pipelines, err
}

// GetPipelineByName gets a Pipeline by name.
func (s *Store) GetPipelineByName(ctx context.Context, name string) (*corev2.Pipeline, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name of pipeline")}
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

// UpdatePipeline updates a Pipeline.
func (s *Store) UpdatePipeline(ctx context.Context, pipeline *corev2.Pipeline) error {
	if err := pipeline.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	return CreateOrUpdate(ctx, s.client, getPipelinePath(pipeline), pipeline.Namespace, pipeline)
}
