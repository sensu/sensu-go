package etcd

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/coreos/etcd/clientv3"
)

type mockBackendIDGetterClient struct {
	grantResp    *clientv3.LeaseGrantResponse
	grantErr     error
	keepaliveCh  chan *clientv3.LeaseKeepAliveResponse
	keepaliveErr error
	putResp      *clientv3.PutResponse
	putErr       error
	puts         []string
	grantCh      chan struct{}
	sync.Mutex
}

func (m *mockBackendIDGetterClient) Grant(ctx context.Context, period int64) (*clientv3.LeaseGrantResponse, error) {
	m.Lock()
	defer m.Unlock()
	defer func() {
		// This is a way to wait for calls to Grant() elsewhere
		m.grantCh <- struct{}{}
	}()
	return m.grantResp, m.grantErr
}

func (m *mockBackendIDGetterClient) KeepAlive(ctx context.Context, id clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	m.Lock()
	defer m.Unlock()
	return m.keepaliveCh, m.keepaliveErr
}

func (m *mockBackendIDGetterClient) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	m.Lock()
	defer m.Unlock()
	m.puts = append(m.puts, fmt.Sprintf("%s=%s", key, val))
	return m.putResp, m.putErr
}

func (m *mockBackendIDGetterClient) clearGrantCh() {
	for {
		select {
		case <-m.grantCh:
		default:
			return
		}
	}
}

func newMockBackendIDGetterClient() *mockBackendIDGetterClient {
	return &mockBackendIDGetterClient{
		grantResp: &clientv3.LeaseGrantResponse{
			ID: clientv3.LeaseID(1234),
		},
		keepaliveCh: make(chan *clientv3.LeaseKeepAliveResponse),
		putResp:     &clientv3.PutResponse{},
		grantCh:     make(chan struct{}, 1000),
	}
}

func TestBackendIDGetter(t *testing.T) {
	client := newMockBackendIDGetterClient()
	getter := NewBackendIDGetter(context.TODO(), client)

	got := getter.GetBackendID()
	if want := int64(1234); got != want {
		t.Fatalf("bad backend id: got %d, want %d", got, want)
	}
}

func TestBackendIDGetterRetry(t *testing.T) {
	// Look, there are some frankly quite concerning concurrency constructs
	// present in this test. I'm sure you aren't thrilled to see them.
	// However they were necessary for me to safely test this thing.
	// Change at your peril :)
	client := newMockBackendIDGetterClient()
	getter := NewBackendIDGetter(context.TODO(), client)

	got := getter.GetBackendID()
	if want := int64(1234); got != want {
		t.Fatalf("bad backend id: got %d, want %d", got, want)
	}

	client.Lock()
	// We need an empty grantCh before progressing
	client.clearGrantCh()

	close(client.keepaliveCh)
	client.grantResp = &clientv3.LeaseGrantResponse{
		ID: clientv3.LeaseID(2345),
	}
	client.keepaliveCh = make(chan *clientv3.LeaseKeepAliveResponse)
	client.Unlock()

	// Wait for Grant() to get called before progressing with the test
	<-client.grantCh

	got = getter.GetBackendID()
	if want := int64(2345); got != want {
		t.Fatalf("bad backend id: got %d, want %d", got, want)
	}
}
