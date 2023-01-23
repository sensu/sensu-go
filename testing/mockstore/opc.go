package mockstore

import (
	"context"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/mock"
)

type OPC struct {
	mock.Mock
}

func (o *OPC) QueryOperator(ctx context.Context, key store.OperatorKey) (store.OperatorState, error) {
	args := o.Called(ctx, key)
	return args.Get(0).(store.OperatorState), args.Error(1)
}

func (o *OPC) ListOperators(ctx context.Context, key store.OperatorKey) ([]store.OperatorState, error) {
	args := o.Called(ctx, key)
	return args.Get(0).([]store.OperatorState), args.Error(1)
}

func (o *OPC) MonitorOperators(ctx context.Context, req store.MonitorOperatorsRequest) <-chan []store.OperatorState {
	args := o.Called(ctx, req)
	return args.Get(0).(<-chan []store.OperatorState)
}

func (o *OPC) CheckIn(ctx context.Context, state store.OperatorState) error {
	return o.Called(ctx, state).Error(0)
}

func (o *OPC) CheckOut(ctx context.Context, key store.OperatorKey) error {
	return o.Called(ctx, key).Error(0)
}
