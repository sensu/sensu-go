// +build integration

package queue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnqueue(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	defer client.Close()
	require.NoError(t, err)

	queue := New("testenq", client, e.BackendID())
	err = queue.Enqueue(context.Background(), "test item")
	assert.NoError(t, err)
}

func TestDequeueSingleItem(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	defer client.Close()
	require.NoError(t, err)

	queue := New("testdeq", client, e.BackendID())
	err = queue.Enqueue(context.Background(), "test single item dequeue")
	require.NoError(t, err)

	item, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test single item dequeue", item.Value())

	defer item.Ack(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	item, err = queue.Dequeue(ctx)
	assert.Error(t, err)
}

func TestDequeueFIFO(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	defer client.Close()
	require.NoError(t, err)

	queue := New("testfifo", client, e.BackendID())
	items := []string{"hello", "there", "world", "asdf", "fjdksl", "lalalal"}

	for _, item := range items {
		require.NoError(t, queue.Enqueue(context.Background(), item))
	}

	result := []string{}

	for range items {
		item, err := queue.Dequeue(context.Background())
		require.NoError(t, err)
		result = append(result, item.Value())
	}

	require.Equal(t, items, result)
}

func TestNack(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	defer client.Close()
	require.NoError(t, err)

	queue := New("testnack", client, e.BackendID())
	err = queue.Enqueue(context.Background(), "test item")
	require.NoError(t, err)

	item, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test item", item.Value())

	err = item.Nack(context.Background())
	require.NoError(t, err)

	item, err = queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test item", item.Value())
}

func TestAck(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	defer client.Close()
	require.NoError(t, err)

	queue := New("testack", client, e.BackendID())
	err = queue.Enqueue(context.Background(), "test item")
	require.NoError(t, err)

	item, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test item", item.Value())

	err = item.Ack(context.Background())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	item, err = queue.Dequeue(ctx)
	require.Error(t, err)
}

func TestOnce(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	defer client.Close()
	require.NoError(t, err)

	queue := New("testonce", client, e.BackendID())

	err = queue.Enqueue(context.Background(), "test item")
	require.NoError(t, err)

	item, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, "test item", item.Value())

	// nack should return the original itemt to the queue for reprocessing, ack
	// called after should have no effect
	require.NoError(t, item.Nack(context.Background()))
	require.NoError(t, item.Ack(context.Background()))
	nackedItem, err := queue.Dequeue(context.Background())
	require.NoError(t, err)
	require.Equal(t, item.Value(), nackedItem.Value())
}

func TestNackExpired(t *testing.T) {
	t.Parallel()
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client, err := e.NewClient()
	defer client.Close()
	require.NoError(t, err)

	queue := New("testexpired", client, e.BackendID())
	queue.itemTimeout = 2 * time.Second

	ctx, cancel := context.WithCancel(context.Background())

	err = queue.Enqueue(ctx, "test item")
	require.NoError(t, err)

	item, err := queue.Dequeue(ctx)
	require.NoError(t, err)

	// close the first client
	err = client.Close()
	require.NoError(t, err)
	cancel()

	// create a new client and queue
	newClient, err := e.NewClient()
	defer newClient.Close()
	require.NoError(t, err)

	// wait to make sure the item has timed out
	time.Sleep(2 * time.Second)

	newQueue := New("testexpired", newClient, e.BackendID())
	newQueue.itemTimeout = 2 * time.Second

	// nacked item should go back in the work queue lane
	item, err = newQueue.Dequeue(context.Background())
	require.NoError(t, err)

	require.Equal(t, "test item", item.Value())
}

type memberListOverrideClient struct {
	*clientv3.Client
	memberListResponse *clientv3.MemberListResponse
	memberListError    error
}

func (o *memberListOverrideClient) MemberList(ctx context.Context) (*clientv3.MemberListResponse, error) {
	return o.memberListResponse, o.memberListError
}

func TestMultipleSubscribers(t *testing.T) {
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	// the override client simulates multiple backends by providing a custom
	// MemberList method
	override := &memberListOverrideClient{
		Client: client,
		memberListResponse: &clientv3.MemberListResponse{
			Header: &etcdserverpb.ResponseHeader{},
			Members: []*etcdserverpb.Member{
				{
					ID: 1,
				},
				{
					ID: 2,
				},
				{
					ID: 3,
				},
			},
		},
	}

	// Each queue is associated with a different backend
	q1 := New("testMultipleSubscribers", override, fmt.Sprintf("%x", 1))
	q2 := New("testMultipleSubscribers", override, fmt.Sprintf("%x", 2))
	q3 := New("testMultipleSubscribers", override, fmt.Sprintf("%x", 3))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	require.NoError(t, q1.Enqueue(ctx, "foobar"))

	// Each queue gets the message!
	for _, q := range []*Queue{q1, q2, q3} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		item, err := q.Dequeue(ctx)
		require.NoError(t, err)
		assert.Equal(t, "foobar", item.Value())
	}
}

func TestCleanupQueue(t *testing.T) {
	// This tests ensures that defunct queues are garbage-collected
	t.Parallel()

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	require.NoError(t, err)
	defer client.Close()

	// the override client simulates multiple backends by providing a custom
	// MemberList method
	memberListResponse := &clientv3.MemberListResponse{
		Header: &etcdserverpb.ResponseHeader{},
		Members: []*etcdserverpb.Member{
			{
				ID: 1,
			},
			{
				ID: 2,
			},
			{
				ID: 3,
			},
		},
	}
	override := &memberListOverrideClient{
		Client:             client,
		memberListResponse: memberListResponse,
	}

	// Each queue is associated with a different backend
	q1 := New("testMultipleSubscribers", override, fmt.Sprintf("%x", 1))
	q2 := New("testMultipleSubscribers", override, fmt.Sprintf("%x", 2))
	q3 := New("testMultipleSubscribers", override, fmt.Sprintf("%x", 3))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// At this point q1, q2 and q3 all have items waiting
	require.NoError(t, q1.Enqueue(ctx, "foobar"))

	// Truncate the member list to be only q1
	oldMembers := memberListResponse.Members
	memberListResponse.Members = oldMembers[:1]

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// After the next enqueue operation, the values for q2 and q3 will be removed
	require.NoError(t, q1.Enqueue(ctx, "barbaz"))

	// Restore the member list to its original state
	memberListResponse.Members = oldMembers

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	require.NoError(t, q1.Enqueue(ctx, "rubber ducky"))

	// q1 expects to get three values
	for _, exp := range []string{"foobar", "barbaz", "rubber ducky"} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		item, err := q1.Dequeue(ctx)
		require.NoError(t, err)
		assert.Equal(t, exp, item.Value())
	}

	// q2 and q3 expect only a single value
	for _, q := range []*Queue{q2, q3} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		item, err := q.Dequeue(ctx)
		require.NoError(t, err)
		assert.Equal(t, "rubber ducky", item.Value())
	}
}
