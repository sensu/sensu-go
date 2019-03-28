// +build integration,!race

package tessend

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/etcd"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTessendTest(t *testing.T) *Tessend {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	cluster := clientv3.NewCluster(client)

	store := etcdstore.NewStore(client, e.Name())

	tessend, err := New(Config{
		Store:   store,
		Cluster: cluster,
		Client:  client,
	})
	require.NoError(t, err)
	return tessend
}

func TestTessend200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("tessen-reporting-interval", "1800")
		assert.NotEmpty(t, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	tessend := newTessendTest(t)
	tessend.url = ts.URL
	require.NoError(t, tessend.Start())
	time.Sleep(3 * time.Second)
	assert.NoError(t, tessend.Stop())
	require.Equal(t, uint32(1800), tessend.interval)
	assert.Equal(t, tessend.Name(), "tessend")
}

func TestTessend200InvalidHeaderKey(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("foo", "1800")
		assert.NotEmpty(t, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	tessend := newTessendTest(t)
	tessend.url = ts.URL
	require.NoError(t, tessend.Start())
	time.Sleep(3 * time.Second)
	assert.NoError(t, tessend.Stop())
	require.Equal(t, uint32(v2.DefaultTessenInterval), tessend.interval)
	assert.Equal(t, tessend.Name(), "tessend")
}

func TestTessend200InvalidHeaderValue(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("tessen-reporting-interval", "216000")
		assert.NotEmpty(t, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	tessend := newTessendTest(t)
	tessend.url = ts.URL
	require.NoError(t, tessend.Start())
	time.Sleep(3 * time.Second)
	assert.NoError(t, tessend.Stop())
	require.Equal(t, uint32(v2.DefaultTessenInterval), tessend.interval)
	assert.Equal(t, tessend.Name(), "tessend")
}

func TestTessend500(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Body)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	tessend := newTessendTest(t)
	tessend.url = ts.URL
	require.NoError(t, tessend.Start())
	time.Sleep(3 * time.Second)
	assert.NoError(t, tessend.Stop())
	assert.Equal(t, tessend.Name(), "tessend")
}

func TestTessendEnabled(t *testing.T) {
	tessen := v2.DefaultTessenConfig()
	tessend := newTessendTest(t)
	require.True(t, tessend.enabled(tessen))
	tessen.OptOut = true
	require.False(t, tessend.enabled(tessen))
}
