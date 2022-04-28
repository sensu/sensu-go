package mockclientv3

import (
	"context"

	"github.com/stretchr/testify/mock"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// MockClientV3 is a V3 Etcd Client used for testing. When using the MockStore in unit
// tests, stub out the behavior you wish to test against by assigning the
// appropriate function to the appropriate Func field. If you have forgotten
// to stub a particular function, the program will panic.
type MockClientV3 struct {
	mock.Mock
}

// Compact ...
func (m MockClientV3) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	args := m.Called(ctx, rev, opts)
	return args.Get(0).(*clientv3.CompactResponse), args.Error(1)
}

// Delete ...
func (m MockClientV3) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	args := m.Called(ctx, key, opts)
	return args.Get(0).(*clientv3.DeleteResponse), args.Error(1)
}

// Do ...
func (m MockClientV3) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	args := m.Called(ctx, op)
	return args.Get(0).(clientv3.OpResponse), args.Error(1)
}

// Get ...
func (m MockClientV3) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	args := m.Called(ctx, key, opts)
	return args.Get(0).(*clientv3.GetResponse), args.Error(1)
}

// Put ...
func (m MockClientV3) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	args := m.Called(ctx, key, val, opts)
	return args.Get(0).(*clientv3.PutResponse), args.Error(1)
}

// Txn ...
func (m MockClientV3) Txn(ctx context.Context) clientv3.Txn {
	args := m.Called(ctx)
	return args.Get(0).(clientv3.Txn)
}
