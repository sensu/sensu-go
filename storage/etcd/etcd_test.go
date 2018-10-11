package etcd_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/runtime/codec"
	"github.com/sensu/sensu-go/storage"
	"github.com/sensu/sensu-go/storage/etcd"
	"github.com/stretchr/testify/mock"
)

type mockKV struct {
	mock.Mock
}

func (m *mockKV) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	args := m.Called(ctx, key, val, opts)
	return args.Get(0).(*clientv3.PutResponse), args.Error(1)
}

func (m *mockKV) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	args := m.Called(ctx, key, opts)
	return args.Get(0).(*clientv3.GetResponse), args.Error(1)
}

func (m *mockKV) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	args := m.Called(ctx, key, opts)
	return args.Get(0).(*clientv3.DeleteResponse), args.Error(1)
}

func (m *mockKV) Compact(ctx context.Context, rev int64, opts ...clientv3.CompactOption) (*clientv3.CompactResponse, error) {
	args := m.Called(ctx, rev, opts)
	return args.Get(0).(*clientv3.CompactResponse), args.Error(1)
}

func (m *mockKV) Do(ctx context.Context, op clientv3.Op) (clientv3.OpResponse, error) {
	args := m.Called(ctx, op)
	return args.Get(0).(clientv3.OpResponse), args.Error(1)
}

func (m *mockKV) Txn(ctx context.Context) clientv3.Txn {
	return m.Called(ctx).Get(0).(clientv3.Txn)
}

type mockTxn struct {
	mock.Mock
}

func (m *mockTxn) If(cs ...clientv3.Cmp) clientv3.Txn {
	return m.Called(cs).Get(0).(clientv3.Txn)
}

func (m *mockTxn) Then(ops ...clientv3.Op) clientv3.Txn {
	return m.Called(ops).Get(0).(clientv3.Txn)
}

func (m *mockTxn) Else(ops ...clientv3.Op) clientv3.Txn {
	return m.Called(ops).Get(0).(clientv3.Txn)
}

func (m *mockTxn) Commit() (*clientv3.TxnResponse, error) {
	args := m.Called()
	return args.Get(0).(*clientv3.TxnResponse), args.Error(1)
}

func TestGetExists(t *testing.T) {
	kv := &mockKV{}
	response := &clientv3.GetResponse{
		Kvs: []*mvccpb.KeyValue{
			{
				Key:   []byte("foo"),
				Value: []byte(`"bar"`),
			},
		},
	}
	kv.On("Get", mock.Anything, "foo", mock.Anything).Return(response, nil)
	store := etcd.NewStorage(kv, codec.UniversalCodec())
	var result string

	if err := store.Get(context.TODO(), "foo", &result); err != nil {
		t.Fatal(err)
	}
	if got, want := result, "bar"; got != want {
		t.Fatalf("bad get: got %q, want %q", got, want)
	}

	kv.AssertCalled(t, "Get", mock.Anything, "foo", mock.Anything)
}

func TestGetNotExists(t *testing.T) {
	kv := &mockKV{}
	response := &clientv3.GetResponse{}
	kv.On("Get", mock.Anything, "foo", mock.Anything).Return(response, nil)
	store := etcd.NewStorage(kv, codec.UniversalCodec())
	var result string

	err := store.Get(context.TODO(), "foo", &result)
	if got, want := err, storage.ErrNotFound; got != want {
		t.Fatalf("bad get error: got %v, want %v", got, want)
	}

	kv.AssertCalled(t, "Get", mock.Anything, "foo", mock.Anything)
}

func TestGetError(t *testing.T) {
	kv := &mockKV{}
	response := (*clientv3.GetResponse)(nil)
	kv.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(response, errors.New("an error"))
	store := etcd.NewStorage(kv, codec.UniversalCodec())
	var result string

	if err := store.Get(context.TODO(), "foo", &result); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestGetCodecError(t *testing.T) {
	kv := &mockKV{}
	response := &clientv3.GetResponse{
		Kvs: []*mvccpb.KeyValue{
			{
				Key:   []byte("foo"),
				Value: []byte(`"bar"`),
			},
		},
	}
	kv.On("Get", mock.Anything, "foo", mock.Anything).Return(response, nil)
	store := etcd.NewStorage(kv, codec.UniversalCodec())
	var result chan struct{}

	if err := store.Get(context.TODO(), "foo", &result); err == nil {
		t.Fatal("expected non-nil error")
	}

	kv.AssertCalled(t, "Get", mock.Anything, "foo", mock.Anything)
}

func TestCreateExists(t *testing.T) {
	kv := &mockKV{}
	txn := &mockTxn{}
	kv.On("Txn", mock.Anything).Return(txn)
	response := &clientv3.TxnResponse{
		Succeeded: false,
	}
	txn.On("Commit").Return(response, nil)
	txn.On("If", mock.Anything).Return(txn)
	txn.On("Then", mock.Anything).Return(txn)
	store := etcd.NewStorage(kv, codec.UniversalCodec())

	if err := store.Create(context.TODO(), "foo", `"bar"`); err == nil {
		t.Fatal("expected non-nil error")
	}

	kv.AssertCalled(t, "Txn", mock.Anything)
	txn.AssertCalled(t, "If", mock.Anything)
	txn.AssertCalled(t, "Then", mock.Anything)
	txn.AssertCalled(t, "Commit")
}

func TestCreateNotExists(t *testing.T) {
	kv := &mockKV{}
	txn := &mockTxn{}
	kv.On("Txn", mock.Anything).Return(txn)
	response := &clientv3.TxnResponse{
		Succeeded: true,
	}
	txn.On("Commit").Return(response, nil)
	txn.On("If", mock.Anything).Return(txn)
	txn.On("Then", mock.Anything).Return(txn)
	store := etcd.NewStorage(kv, codec.UniversalCodec())

	if err := store.Create(context.TODO(), "foo", `"bar"`); err != nil {
		t.Fatal(err)
	}

	kv.AssertCalled(t, "Txn", mock.Anything)
	txn.AssertCalled(t, "If", mock.Anything)
	txn.AssertCalled(t, "Then", mock.Anything)
	txn.AssertCalled(t, "Commit")
}

func TestCreateCodecError(t *testing.T) {
	kv := &mockKV{}
	store := etcd.NewStorage(kv, codec.UniversalCodec())
	if err := store.Create(context.TODO(), "foo", make(chan struct{})); err == nil {
		t.Fatal("expected non-nil error")
	}
	kv.AssertNotCalled(t, "Txn", mock.Anything)
}

func TestCreateKVError(t *testing.T) {
	kv := &mockKV{}
	txn := &mockTxn{}
	kv.On("Txn", mock.Anything).Return(txn)
	response := (*clientv3.TxnResponse)(nil)
	txn.On("Commit").Return(response, errors.New("an error"))
	txn.On("If", mock.Anything).Return(txn)
	txn.On("Then", mock.Anything).Return(txn)
	store := etcd.NewStorage(kv, codec.UniversalCodec())

	if err := store.Create(context.TODO(), "foo", `"bar"`); err == nil {
		t.Fatal("expected non-nil error")
	}

	kv.AssertCalled(t, "Txn", mock.Anything)
	txn.AssertCalled(t, "If", mock.Anything)
	txn.AssertCalled(t, "Then", mock.Anything)
	txn.AssertCalled(t, "Commit")
}

func TestCreateOrUpdateSuccess(t *testing.T) {
	kv := &mockKV{}
	response := &clientv3.PutResponse{}
	kv.On("Put", mock.Anything, "foo", `"bar"`, mock.Anything).Return(response, nil)
	store := etcd.NewStorage(kv, codec.UniversalCodec())
	if err := store.CreateOrUpdate(context.TODO(), "foo", "bar"); err != nil {
		t.Fatal(err)
	}
	kv.AssertCalled(t, "Put", mock.Anything, "foo", `"bar"`, mock.Anything)
}

func TestCreateOrUpdateError(t *testing.T) {
	kv := &mockKV{}
	response := &clientv3.PutResponse{}
	kv.On("Put", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(response, errors.New("an error"))
	store := etcd.NewStorage(kv, codec.UniversalCodec())
	if err := store.CreateOrUpdate(context.TODO(), "foo", "bar"); err == nil {
		t.Fatal("expected non-nil error")
	}
	kv.AssertCalled(t, "Put", mock.Anything, "foo", `"bar"`, mock.Anything)
}

func TestListSuccess(t *testing.T) {
	kv := &mockKV{}
	response := &clientv3.GetResponse{
		Kvs: []*mvccpb.KeyValue{
			{
				Value: []byte(`"bar"`),
			},
			{
				Value: []byte(`"foo"`),
			},
		},
	}
	kv.On("Get", mock.Anything, "foo", mock.Anything).Return(response, nil)
	store := etcd.NewStorage(kv, codec.UniversalCodec())
	var result []string
	if err := store.List(context.TODO(), "foo", &result); err != nil {
		t.Fatal(err)
	}
	if got, want := result, []string{"bar", "foo"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad list: got %v, want %v", got, want)
	}

	kv.AssertCalled(t, "Get", mock.Anything, "foo", mock.Anything)
}

func TestListFailure(t *testing.T) {
	kv := &mockKV{}
	response := (*clientv3.GetResponse)(nil)
	kv.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(response, errors.New("an error"))

	store := etcd.NewStorage(kv, codec.UniversalCodec())
	if err := store.List(context.TODO(), "foo", &[]string{}); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestListBadResultParam(t *testing.T) {
	kv := &mockKV{}
	store := etcd.NewStorage(kv, codec.UniversalCodec())
	if err := store.List(context.TODO(), "foo", new(int)); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestUpdateExists(t *testing.T) {
	kv := &mockKV{}
	txn := &mockTxn{}
	kv.On("Txn", mock.Anything).Return(txn)
	response := &clientv3.TxnResponse{
		Succeeded: true,
	}
	txn.On("Commit").Return(response, nil)
	txn.On("If", mock.Anything).Return(txn)
	txn.On("Then", mock.Anything).Return(txn)

	store := etcd.NewStorage(kv, codec.UniversalCodec())

	if err := store.Update(context.TODO(), "foo", `"bar"`); err != nil {
		t.Fatal(err)
	}

	kv.AssertCalled(t, "Txn", mock.Anything)
	txn.AssertCalled(t, "If", mock.Anything)
	txn.AssertCalled(t, "Then", mock.Anything)
	txn.AssertCalled(t, "Commit")
}

func TestUpdateNotExists(t *testing.T) {
	kv := &mockKV{}
	txn := &mockTxn{}
	kv.On("Txn", mock.Anything).Return(txn)
	response := &clientv3.TxnResponse{
		Succeeded: false,
	}
	txn.On("Commit").Return(response, nil)
	txn.On("If", mock.Anything).Return(txn)
	txn.On("Then", mock.Anything).Return(txn)

	store := etcd.NewStorage(kv, codec.UniversalCodec())

	if err := store.Update(context.TODO(), "foo", `"bar"`); err == nil {
		t.Fatal("expected non-nil error")
	}

	kv.AssertCalled(t, "Txn", mock.Anything)
	txn.AssertCalled(t, "If", mock.Anything)
	txn.AssertCalled(t, "Then", mock.Anything)
	txn.AssertCalled(t, "Commit")
}

func TestUpdateCodecError(t *testing.T) {
	kv := &mockKV{}
	store := etcd.NewStorage(kv, codec.UniversalCodec())
	if err := store.Update(context.TODO(), "foo", make(chan struct{})); err == nil {
		t.Fatal("expected non-nil error")
	}
	kv.AssertNotCalled(t, "Txn", mock.Anything)
}
