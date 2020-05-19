// +build integration,!race

package tessend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTessendTest(t *testing.T) *Tessend {
	bus, err := messaging.NewWizardBus(messaging.WizardBusConfig{})
	require.NoError(t, err)
	require.NoError(t, bus.Start())

	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()
	defer client.Close()

	s := &mockstore.MockStore{}
	ch := make(<-chan store.WatchEventTessenConfig)
	s.On("CreateOrUpdateTessenConfig", mock.Anything, mock.Anything).Return(fmt.Errorf("foo"))
	s.On("GetTessenConfig", mock.Anything, mock.Anything).Return(corev2.DefaultTessenConfig(), fmt.Errorf("foo"))
	s.On("GetTessenConfigWatcher", mock.Anything).Return(ch)
	s.On("GetClusterID", mock.Anything).Return("foo", fmt.Errorf("foo"))

	tessend, err := New(
		context.Background(),
		Config{
			Store:    s,
			Client:   client,
			RingPool: ringv2.NewPool(client),
			Bus:      bus,
		})
	tessend.duration = 5 * time.Millisecond
	tessend.config = corev2.DefaultTessenConfig()
	require.NoError(t, err)
	return tessend
}

func TestTessendBus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	tessend := newTessendTest(t)
	tessend.url = ts.URL
	require.NoError(t, tessend.Start())
	require.NoError(t, tessend.bus.Publish(messaging.TopicTessen, corev2.DefaultTessenConfig()))
	time.Sleep(2 * time.Second)
	assert.NoError(t, tessend.Stop())
	assert.Equal(t, tessend.Name(), "tessend")
}

func TestTessendBusMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.NotEmpty(t, r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	tessend := newTessendTest(t)
	tessend.url = ts.URL
	require.NoError(t, tessend.Start())
	require.NoError(t, tessend.bus.Publish(messaging.TopicTessenMetric, []corev2.MetricPoint{
		corev2.MetricPoint{
			Name:  "metric",
			Value: 1,
		},
	}))
	time.Sleep(2 * time.Second)
	assert.NoError(t, tessend.Stop())
	assert.Equal(t, tessend.Name(), "tessend")
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
	require.Equal(t, uint32(corev2.DefaultTessenInterval), tessend.interval)
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
	require.Equal(t, uint32(corev2.DefaultTessenInterval), tessend.interval)
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
	tessen := corev2.DefaultTessenConfig()
	tessend := newTessendTest(t)
	require.True(t, tessend.enabled())
	tessen.OptOut = true
	tessend.config = tessen
	require.False(t, tessend.enabled())
}

func TestInternalTag(t *testing.T) {
	mp := &corev2.MetricPoint{
		Name:  "test_metric",
		Value: 5,
	}
	os.Clearenv()
	appendInternalTag(mp)
	assert.Equal(t, 0, len(mp.Tags))
	os.Setenv("SENSU_INTERNAL_ENVIRONMENT", "foo")
	appendInternalTag(mp)
	assert.Equal(t, 1, len(mp.Tags))
	assert.Equal(t, "sensu_internal_environment", mp.Tags[0].Name)
	assert.Equal(t, "foo", mp.Tags[0].Value)
}

func TestGetEntityClassCounts(t *testing.T) {
	tessend := newTessendTest(t)
	now := time.Now().Unix()

	// default entity class counts func
	points := tessend.getEntityClassMetrics(now)
	assert.Equal(t, 0, len(points))

	// override entity class counts func
	tessend.EntityClassCounts = func() map[string]int {
		counts := make(map[string]int)
		counts["foo"] = 10
		counts["bar"] = 20
		return counts
	}

	storeCfg := tessend.GetStoreConfig()
	expectedFoo := &corev2.MetricPoint{
		Name:      "entity_class_foo_count",
		Value:     float64(10),
		Timestamp: now,
	}
	appendInternalTag(expectedFoo)
	appendStoreConfig(expectedFoo, storeCfg)

	expectedBar := &corev2.MetricPoint{
		Name:      "entity_class_bar_count",
		Value:     float64(20),
		Timestamp: now,
	}
	appendInternalTag(expectedBar)
	appendStoreConfig(expectedBar, storeCfg)

	points = tessend.getEntityClassMetrics(now)
	assert.Equal(t, 2, len(points))
	assert.Contains(t, points, expectedFoo)
	assert.Contains(t, points, expectedBar)
}
