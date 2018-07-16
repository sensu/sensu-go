package graphql

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

// checks

type mockCheckExecutor struct {
	err error
}

func (m mockCheckExecutor) QueueAdhocRequest(_ context.Context, _ string, _ *types.AdhocRequest) error {
	return m.err
}

// entities

type mockEntityFetcher struct {
	record *types.Entity
	err    error
}

func (m mockEntityFetcher) Find(_ context.Context, _ string) (*types.Entity, error) {
	return m.record, m.err
}

type mockEntityDestroyer struct {
	err error
}

func (m mockEntityDestroyer) Destroy(_ context.Context, _ string) error {
	return m.err
}

// events

type mockEventQuerier struct {
	els []*types.Event
	err error
}

func (f mockEventQuerier) Query(ctx context.Context, entity, check string) ([]*types.Event, error) {
	return f.els, f.err
}

type mockEventFetcher struct {
	record *types.Event
	err    error
}

func (m mockEventFetcher) Find(ctx context.Context, entity, check string) (*types.Event, error) {
	return m.record, m.err
}

type mockEventDestroyer struct {
	err error
}

func (m mockEventDestroyer) Destroy(ctx context.Context, a, b string) error {
	return m.err
}

// environments

type mockEnvironmentFinder struct {
	record *types.Environment
	err    error
}

func (m mockEnvironmentFinder) Find(_ context.Context, org, env string) (*types.Environment, error) {
	if org != "bobs-burgers" || env != "us-west-2" {
		return nil, nil
	}
	return m.record, m.err
}

// silences

type mockSilenceCreator struct {
	err error
}

func (m mockSilenceCreator) Create(_ context.Context, _ *types.Silenced) error {
	return m.err
}

type mockSilenceDestroyer struct {
	err error
}

func (m mockSilenceDestroyer) Destroy(_ context.Context, _ string) error {
	return m.err
}

type mockSilenceQuerier struct {
	els []*types.Silenced
	err error
}

func (m mockSilenceQuerier) Query(_ context.Context, _, _ string) ([]*types.Silenced, error) {
	return m.els, m.err
}
