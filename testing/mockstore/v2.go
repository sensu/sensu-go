package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/stretchr/testify/mock"
)

type V2MockStore struct {
	mock.Mock
}

func (v *V2MockStore) CreateOrUpdate(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper) error {
	return v.Called(ctx, req, w).Error(0)
}

func (v *V2MockStore) UpdateIfExists(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper) error {
	return v.Called(ctx, req, w).Error(0)
}

func (v *V2MockStore) CreateIfNotExists(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper) error {
	return v.Called(ctx, req, w).Error(0)
}

func (v *V2MockStore) Get(ctx context.Context, req storev2.ResourceRequest) (storev2.Wrapper, error) {
	args := v.Called(ctx, req)
	wrapper, _ := args.Get(0).(storev2.Wrapper)
	return wrapper, args.Error(1)
}

func (v *V2MockStore) Delete(ctx context.Context, req storev2.ResourceRequest) error {
	return v.Called(ctx, req).Error(0)
}

func (v *V2MockStore) List(ctx context.Context, req storev2.ResourceRequest, pred *store.SelectionPredicate) (storev2.WrapList, error) {
	args := v.Called(ctx, req, pred)
	list, _ := args.Get(0).(wrap.List)
	return list, args.Error(1)
}

func (v *V2MockStore) Exists(ctx context.Context, req storev2.ResourceRequest) (bool, error) {
	args := v.Called(ctx, req)
	return args.Get(0).(bool), args.Error(1)
}

func (v *V2MockStore) Patch(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper, patcher patch.Patcher, cond *store.ETagCondition) error {
	return v.Called(ctx, req, w, patcher, cond).Error(0)
}
