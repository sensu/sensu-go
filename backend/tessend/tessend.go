package tessend

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/google/uuid"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/ringv2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/version"
	"github.com/sirupsen/logrus"
)

const (
	// tessenURL is the http endpoint for the tessen service.
	tessenURL = "https://tessen.sensu.io/v2/data"

	// ringUpdateInterval is the interval, in seconds, that TessenD will
	// update the ring with any added/removed cluster members.
	ringUpdateInterval = 450 * time.Second

	// ringBackendKeepalive is the length of time, in seconds, that the
	// ring considers an entry alive.
	ringBackendKeepalive = 900
)

// Tessend is the tessen daemon.
type Tessend struct {
	interval  uint32
	store     store.Store
	ctx       context.Context
	cancel    context.CancelFunc
	errChan   chan error
	ring      *ringv2.Ring
	interrupt chan *corev2.TessenConfig
	client    *clientv3.Client
	url       string
	backendID string
}

// Option is a functional option.
type Option func(*Tessend) error

// Config configures Tessend.
type Config struct {
	Store    store.Store
	RingPool *ringv2.Pool
	Client   *clientv3.Client
}

// New creates a new TessenD.
func New(c Config, opts ...Option) (*Tessend, error) {
	t := &Tessend{
		interval:  corev2.DefaultTessenInterval,
		store:     c.Store,
		client:    c.Client,
		errChan:   make(chan error, 1),
		url:       tessenURL,
		backendID: uuid.New().String(),
	}
	t.ctx, t.cancel = context.WithCancel(context.Background())
	t.interrupt = make(chan *corev2.TessenConfig, 1)
	key := ringv2.Path("global", "backends")
	t.ring = c.RingPool.Get(key)

	return t, nil
}

// Start the Tessen daemon.
func (t *Tessend) Start() error {
	tessen, err := t.store.GetTessenConfig(t.ctx)
	// create the default tessen config if one does not already exist
	if err != nil || tessen == nil {
		tessen = corev2.DefaultTessenConfig()
		err = t.store.CreateOrUpdateTessenConfig(t.ctx, tessen)
		if err != nil {
			// log the error and continue with the default config
			logger.WithError(err).Error("unable to update tessen store")
		}
	}

	if err := t.ctx.Err(); err != nil {
		return err
	}

	go t.startWatcher()
	go t.startRingUpdates()
	go t.start(tessen)
	// Attempt to send data immediately if tessen is enabled
	if t.enabled(tessen) {
		t.collectAndSend(tessen)
	}

	return nil
}

// Stop the Tessen daemon.
func (t *Tessend) Stop() error {
	if err := t.ring.Remove(t.ctx, t.backendID); err != nil {
		logger.WithField("key", t.backendID).WithError(err).Error("error removing key from the ring")
	} else {
		logger.WithField("key", t.backendID).Debug("removed a key from the ring")
	}
	t.cancel()
	close(t.errChan)
	return nil
}

// Err returns a channel on which to listen for terminal errors.
func (t *Tessend) Err() <-chan error {
	return t.errChan
}

// Name returns the daemon name.
func (t *Tessend) Name() string {
	return "tessend"
}

// startWatcher watches the TessenConfig store for changes to the opt-out configuration.
func (t *Tessend) startWatcher() {
	watchChan := t.store.GetTessenConfigWatcher(t.ctx)
	for {
		select {
		case watchEvent, ok := <-watchChan:
			if !ok {
				// The watchChan has closed. Restart the watcher.
				watchChan = t.store.GetTessenConfigWatcher(t.ctx)
				continue
			}
			t.handleWatchEvent(watchEvent)
		case <-t.ctx.Done():
			return
		}
	}
}

// handleWatchEvent issues an interrupt if a change to the stored TessenConfig has been detected.
func (t *Tessend) handleWatchEvent(watchEvent store.WatchEventTessenConfig) {
	tessen := watchEvent.TessenConfig

	switch watchEvent.Action {
	case store.WatchCreate:
		logger.WithField("opt-out", tessen.OptOut).Debug("tessen configuration created")
	case store.WatchUpdate:
		logger.WithField("opt-out", tessen.OptOut).Debug("tessen configuration updated")
	case store.WatchDelete:
		logger.WithField("opt-out", tessen.OptOut).Debug("tessen configuration deleted")
	}

	t.interrupt <- tessen
}

// startRingUpdates starts a loop to periodically update the ring.
func (t *Tessend) startRingUpdates() {
	ticker := time.NewTicker(ringUpdateInterval)
	defer ticker.Stop()
	t.updateRing()
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			t.updateRing()
		}
	}
}

// updateRing adds/updates the ring with a given key.
func (t *Tessend) updateRing() {
	if err := t.ring.Add(t.ctx, t.backendID, ringBackendKeepalive); err != nil {
		logger.WithField("key", t.backendID).WithError(err).Error("error adding key to the ring")
	} else {
		logger.WithField("key", t.backendID).Debug("added a key to the ring")
	}
}

// watchRing watches the ring and handles ring events. It recreates watchers
// when they terminate due to error.
func (t *Tessend) watchRing(ctx context.Context, tessen *corev2.TessenConfig, wg *sync.WaitGroup) {
	wc := t.ring.Watch(ctx, "tessen", 1, int(t.interval), "")
	go func() {
		t.handleEvents(tessen, wc)
		defer wg.Done()
	}()
}

// handleEvents logs different ring events and triggers tessen to run if applicable.
func (t *Tessend) handleEvents(tessen *corev2.TessenConfig, ch <-chan ringv2.Event) {
	for event := range ch {
		switch event.Type {
		case ringv2.EventError:
			logger.WithError(event.Err).Error("ring event error")
		case ringv2.EventAdd:
			logger.WithField("values", event.Values).Debug("added a backend to tessen")
		case ringv2.EventRemove:
			logger.WithField("values", event.Values).Debug("removed a backend from tessen")
		case ringv2.EventTrigger:
			logger.WithField("values", event.Values).Debug("tessen ring trigger")
			// only trigger tessen if the next backend in the ring is this backend
			if event.Values[0] == t.backendID {
				if t.enabled(tessen) {
					t.collectAndSend(tessen)
				}
			}
		case ringv2.EventClosing:
			logger.Debug("tessen ring closing")
		}
	}
}

// start starts the tessen service.
func (t *Tessend) start(tessen *corev2.TessenConfig) {
	ctx, cancel := context.WithCancel(t.ctx)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	t.watchRing(ctx, tessen, wg)

	for {
		select {
		case <-t.ctx.Done():
			cancel()
			return
		case config := <-t.interrupt:
			// Config change indicates the need to recreate the watcher
			cancel()
			wg.Wait()
			ctx, cancel = context.WithCancel(t.ctx)
			wg.Add(1)
			t.watchRing(ctx, config, wg)
		}
	}
}

// enabled checks the tessen config for opt-out status, and verifies the existence of an enterprise license.
// It returns a boolean value indicating if tessen should be enabled or not.
func (t *Tessend) enabled(tessen *corev2.TessenConfig) bool {
	if !tessen.OptOut {
		logger.WithField("opt-out", tessen.OptOut).Info("tessen is opted in, enabling tessen.. thank you so much for your support 💚")
		return true
	}

	wrapper := &Wrapper{}
	err := etcd.Get(t.ctx, t.client, licenseStorePath, wrapper)
	if err != nil {
		logger.WithField("opt-out", tessen.OptOut).Info("tessen is opted out, patiently waiting for you to opt back in")
	} else {
		logger.WithField("opt-out", tessen.OptOut).Info("tessen is opted out but a enterprise license is detected, enabling tessen.. thank you so much for your support 💚")
	}

	return err == nil
}

// collectAndSend is a durable function to collect and send data to tessen.
// Errors are logged and tessen continues to the best of its ability.
func (t *Tessend) collectAndSend(tessen *corev2.TessenConfig) {
	// collect data
	data := t.collect(time.Now().UTC().Unix())

	logger.WithFields(logrus.Fields{
		"url":                       t.url,
		"id":                        data.Cluster.ID,
		data.Metrics.Points[0].Name: data.Metrics.Points[0].Value,
		data.Metrics.Points[1].Name: data.Metrics.Points[1].Value,
	}).Info("sending data to tessen")

	// send data
	resp, err := t.send(data)
	if err != nil {
		logger.WithError(err).Error("tessen phone-home service failed")
		return
	}
	if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		logger.Errorf("bad status: %d (%q)", resp.StatusCode, string(body))
		return
	}

	// parse the response header for an integer value
	interval, err := strconv.ParseUint(resp.Header.Get("tessen-reporting-interval"), 10, 32)
	if err != nil {
		logger.Debugf("invalid tessen response header: %v", err)
		return
	}

	// validate the returned interval is within the upper/lower bound limits
	err = corev2.ValidateInterval(uint32(interval))
	if err != nil {
		logger.Debugf("invalid tessen response header: %v", err)
		return
	}

	// update the tessen interval if the response header returns a new value
	if t.interval != uint32(interval) {
		t.interval = uint32(interval)
		logger.WithField("interval", t.interval).Debug("tessen interval updated")
		t.interrupt <- tessen
	}
}

// collect data and populate the data payload
func (t *Tessend) collect(now int64) *Data {
	var clusterID string
	var entityCount, backendCount float64

	// collect client count
	entities, err := t.store.GetEntities(t.ctx, &store.SelectionPredicate{})
	if err != nil {
		logger.WithError(err).Error("unable to retrieve client count")
	}
	if entities != nil {
		entityCount = float64(len(entities))
	}

	// collect server count and cluster id
	servers, err := t.client.Cluster.MemberList(t.ctx)
	if err != nil {
		logger.WithError(err).Error("unable to retrieve cluster information")
	}
	if servers != nil {
		clusterID = fmt.Sprintf("%x", servers.Header.ClusterId)
		backendCount = float64(len(servers.Members))
	}

	// collect license information
	wrapper := &Wrapper{}
	err = etcd.Get(t.ctx, t.client, licenseStorePath, wrapper)
	if err != nil {
		logger.Debugf("cannot retrieve license: %v", err)
	}

	// populate data payload
	data := &Data{
		Cluster: Cluster{
			ID:      clusterID,
			Version: version.Semver(),
			License: wrapper.Value.License,
		},
		Metrics: corev2.Metrics{
			Points: []*corev2.MetricPoint{
				&corev2.MetricPoint{
					Name:      "entity_count",
					Value:     entityCount,
					Timestamp: now,
				},
				&corev2.MetricPoint{
					Name:      "backend_count",
					Value:     backendCount,
					Timestamp: now,
				},
			},
		},
	}

	return data
}

// send the data payload to the tessen url
func (t *Tessend) send(data *Data) (*http.Response, error) {
	b, _ := json.Marshal(data)
	return http.Post(t.url, "application/json", bytes.NewBuffer(b))
}
