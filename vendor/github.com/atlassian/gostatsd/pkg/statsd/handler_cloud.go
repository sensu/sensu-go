package statsd

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/atlassian/gostatsd"
	stats "github.com/atlassian/gostatsd/pkg/statser"

	"github.com/ash2k/stager/wait"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type lookupResult struct {
	err      error
	ip       gostatsd.IP
	instance *gostatsd.Instance // Can be nil if lookup failed or instance was not found
}

type instanceHolder struct {
	lastAccessNano int64
	expires        time.Time          // When this record expires.
	instance       *gostatsd.Instance // Can be nil if the lookup resulted in an error or instance was not found
}

func (ih *instanceHolder) updateAccess() {
	atomic.StoreInt64(&ih.lastAccessNano, time.Now().UnixNano())
}

func (ih *instanceHolder) lastAccess() int64 {
	return atomic.LoadInt64(&ih.lastAccessNano)
}

// CacheOptions holds cache behaviour configuration.
type CacheOptions struct {
	CacheRefreshPeriod        time.Duration
	CacheEvictAfterIdlePeriod time.Duration
	CacheTTL                  time.Duration
	CacheNegativeTTL          time.Duration
}

// CloudHandler enriches metrics and events with additional information fetched from cloud provider.
type CloudHandler struct {
	// statsCacheHit is accessed by any go routine, must use atomic ops
	statsCacheHit uint64 // Cumulative number of cache hits

	// All other stats fields may only be read or written by the main CloudHandler.Run goroutine
	statsCacheLateHit         uint64 // Cumulative number of late cache hits
	statsCacheMiss            uint64 // Cumulative number of cache misses
	statsCachePositive        uint64 // Absolute number of positive entries in cache
	statsCacheNegative        uint64 // Absolute number of negative entries in cache
	statsCacheRefreshPositive uint64 // Cumulative number of positive refreshes (ie, a refresh which succeeded)
	statsCacheRefreshNegative uint64 // Cumulative number of negative refreshes (ie, a refresh which failed and used old data)
	statsMetricItemsQueued    uint64 // Absolute number of metrics queued, waiting for a CP to respond
	statsMetricHostsQueued    uint64 // Absolute number of IPs waiting for a CP to respond for metrics
	statsEventItemsQueued     uint64 // Absolute number of events queued, waiting for a CP to respond
	statsEventHostsQueued     uint64 // Absolute number of IPs waiting for a CP to respond for events

	// emitChan triggers a write of all the current stats when it is given a Statser
	emitChan chan stats.Statser

	cacheOpts       CacheOptions
	cloud           gostatsd.CloudProvider // Cloud provider interface
	metrics         MetricHandler
	events          EventHandler
	limiter         *rate.Limiter
	metricSource    chan *gostatsd.Metric
	eventSource     chan *gostatsd.Event
	awaitingEvents  map[gostatsd.IP][]*gostatsd.Event
	awaitingMetrics map[gostatsd.IP][]*gostatsd.Metric
	toLookupIPs     []gostatsd.IP
	wg              sync.WaitGroup

	rw    sync.RWMutex // Protects cache
	cache map[gostatsd.IP]*instanceHolder

	estimatedTags int
}

// NewCloudHandler initialises a new cloud handler.
func NewCloudHandler(cloud gostatsd.CloudProvider, metrics MetricHandler, events EventHandler, limiter *rate.Limiter, cacheOptions *CacheOptions) *CloudHandler {
	return &CloudHandler{
		cacheOpts:       *cacheOptions,
		cloud:           cloud,
		metrics:         metrics,
		events:          events,
		limiter:         limiter,
		metricSource:    make(chan *gostatsd.Metric),
		eventSource:     make(chan *gostatsd.Event),
		emitChan:        make(chan stats.Statser),
		awaitingEvents:  make(map[gostatsd.IP][]*gostatsd.Event),
		awaitingMetrics: make(map[gostatsd.IP][]*gostatsd.Metric),
		cache:           make(map[gostatsd.IP]*instanceHolder),
		estimatedTags:   metrics.EstimatedTags() + cloud.EstimatedTags(),
	}
}

// EstimatedTags returns a guess for how many tags to pre-allocate
func (ch *CloudHandler) EstimatedTags() int {
	return ch.estimatedTags
}

func (ch *CloudHandler) DispatchMetric(ctx context.Context, m *gostatsd.Metric) error {
	if ch.updateTagsAndHostname(m.SourceIP, &m.Tags, &m.Hostname) {
		atomic.AddUint64(&ch.statsCacheHit, 1)
		return ch.metrics.DispatchMetric(ctx, m)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case ch.metricSource <- m:
		return nil
	}
}

func (ch *CloudHandler) DispatchEvent(ctx context.Context, e *gostatsd.Event) error {
	if ch.updateTagsAndHostname(e.SourceIP, &e.Tags, &e.Hostname) {
		atomic.AddUint64(&ch.statsCacheHit, 1)
		return ch.events.DispatchEvent(ctx, e)
	}
	ch.wg.Add(1) // Increment before sending to the channel
	select {
	case <-ctx.Done():
		ch.wg.Done()
		return ctx.Err()
	case ch.eventSource <- e:
		return nil
	}
}

// WaitForEvents waits for all event-dispatching goroutines to finish.
func (ch *CloudHandler) WaitForEvents() {
	ch.wg.Wait()
	ch.events.WaitForEvents()
}

func (ch *CloudHandler) RunMetrics(ctx context.Context, statser stats.Statser) {
	if me, ok := ch.cloud.(MetricEmitter); ok {
		var wg wait.Group
		defer wg.Wait()
		wg.Start(func() {
			me.RunMetrics(ctx, statser)
		})
	}

	// All the channels are unbuffered, so no CSWs
	flushed, unregister := statser.RegisterFlush()
	defer unregister()

	for {
		select {
		case <-ctx.Done():
			return
		case <-flushed:
			ch.scheduleEmit(ctx, statser)
		}
	}
}

// scheduleEmit is used to push a request to the main goroutine requesting metrics
// be emitted.  This is done so we can skip atomic operations on most of our metric
// counters.  In line with the flush notifier, it is fire and forget and won't block
func (ch *CloudHandler) scheduleEmit(ctx context.Context, statser stats.Statser) {
	select {
	case ch.emitChan <- statser:
		// success
	case <-ctx.Done():
		// success-ish
	default:
		// at least we tried
	}
}

func (ch *CloudHandler) emit(statser stats.Statser) {
	// atomic
	statser.Gauge("cloudprovider.cache_hit", float64(atomic.LoadUint64(&ch.statsCacheHit)), nil)
	// regular
	statser.Gauge("cloudprovider.cache_late_hit", float64(ch.statsCacheLateHit), nil)
	statser.Gauge("cloudprovider.cache_miss", float64(ch.statsCacheMiss), nil)
	statser.Gauge("cloudprovider.cache_positive", float64(ch.statsCachePositive), nil)
	statser.Gauge("cloudprovider.cache_negative", float64(ch.statsCacheNegative), nil)
	statser.Gauge("cloudprovider.cache_refresh_positive", float64(ch.statsCacheRefreshPositive), nil)
	statser.Gauge("cloudprovider.cache_refresh_negative", float64(ch.statsCacheRefreshNegative), nil)
	t := gostatsd.Tags{"type:metric"}
	statser.Gauge("cloudprovider.hosts_queued", float64(ch.statsMetricHostsQueued), t)
	statser.Gauge("cloudprovider.items_queued", float64(ch.statsMetricItemsQueued), t)
	t = gostatsd.Tags{"type:event"}
	statser.Gauge("cloudprovider.hosts_queued", float64(ch.statsEventHostsQueued), t)
	statser.Gauge("cloudprovider.items_queued", float64(ch.statsEventItemsQueued), t)
}

func (ch *CloudHandler) Run(ctx context.Context) {
	// IPs to lookup. Can make the channel bigger or smaller but this is the perfect size.
	toLookup := make(chan gostatsd.IP, ch.cloud.MaxInstancesBatch())
	var toLookupC chan<- gostatsd.IP
	var toLookupIP gostatsd.IP

	lookupResults := make(chan *lookupResult)

	ld := lookupDispatcher{
		limiter:       ch.limiter,
		cloud:         ch.cloud,
		toLookup:      toLookup,
		lookupResults: lookupResults,
	}

	var wg wait.Group
	defer wg.Wait() // Wait for lookupDispatcher to stop

	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // Tell lookupDispatcher to stop

	wg.StartWithContext(ctx, ld.run)

	refreshTicker := time.NewTicker(ch.cacheOpts.CacheRefreshPeriod)
	defer refreshTicker.Stop()
	// No locking for ch.cache READ access required - this goroutine owns the object and only it mutates it.
	// So reads from the same goroutine are always safe (no concurrent mutations).
	// When we mutate the cache, we hold the exclusive (write) lock to avoid concurrent reads.
	// When we read from the cache from other goroutines, we obtain the read lock.
	for {
		select {
		case <-ctx.Done():
			return
		case toLookupC <- toLookupIP:
			toLookupIP = gostatsd.UnknownIP
			toLookupC = nil // ip has been sent; if there is nothing to send, will block
		case lr := <-lookupResults:
			ch.handleLookupResult(ctx, lr)
		case t := <-refreshTicker.C:
			ch.doRefresh(ctx, t)
		case m := <-ch.metricSource:
			ch.handleMetric(ctx, m)
		case e := <-ch.eventSource:
			ch.handleEvent(ctx, e)
		case statser := <-ch.emitChan:
			ch.emit(statser)
		}
		if toLookupC == nil && len(ch.toLookupIPs) > 0 {
			last := len(ch.toLookupIPs) - 1
			toLookupIP = ch.toLookupIPs[last]
			ch.toLookupIPs[last] = gostatsd.UnknownIP // Enable GC
			ch.toLookupIPs = ch.toLookupIPs[:last]
			toLookupC = toLookup
		}
	}
}

func (ch *CloudHandler) doRefresh(ctx context.Context, t time.Time) {
	var toDelete []gostatsd.IP
	now := t.UnixNano()
	idleNano := ch.cacheOpts.CacheEvictAfterIdlePeriod.Nanoseconds()

	for ip, holder := range ch.cache {
		if now-holder.lastAccess() > idleNano {
			// Entry was not used recently, remove it.
			toDelete = append(toDelete, ip)
			if holder.instance == nil {
				ch.statsCacheNegative--
			} else {
				ch.statsCachePositive--
			}
		} else if t.After(holder.expires) {
			// Entry needs a refresh.
			ch.toLookupIPs = append(ch.toLookupIPs, ip)
		}
	}

	if len(toDelete) > 0 {
		ch.rw.Lock()
		defer ch.rw.Unlock()
		for _, ip := range toDelete {
			delete(ch.cache, ip)
		}
	}
}

func (ch *CloudHandler) handleLookupResult(ctx context.Context, lr *lookupResult) {
	var ttl time.Duration
	if lr.err != nil {
		log.Infof("Error retrieving instance details from cloud provider for %s: %v", lr.ip, lr.err)
		ttl = ch.cacheOpts.CacheNegativeTTL
	} else {
		ttl = ch.cacheOpts.CacheTTL
	}
	now := time.Now()
	newHolder := &instanceHolder{
		expires:  now.Add(ttl),
		instance: lr.instance,
	}
	currentHolder := ch.cache[lr.ip]
	if currentHolder == nil {
		// Not in cache, count it
		if lr.err != nil {
			ch.statsCacheNegative++
		} else {
			ch.statsCachePositive++
		}
		newHolder.lastAccessNano = now.UnixNano()
	} else {
		// In cache, don't count it
		newHolder.lastAccessNano = currentHolder.lastAccess()
		if lr.err != nil {
			// Use the old instance if there was a lookup error.
			newHolder.instance = currentHolder.instance
			ch.statsCacheRefreshNegative++
		} else {
			if currentHolder.instance == nil && newHolder.instance != nil {
				// An entry has flipped from invalid to valid
				ch.statsCacheNegative--
				ch.statsCachePositive++
			}
			ch.statsCacheRefreshPositive++
		}
	}
	func() {
		ch.rw.Lock()
		defer ch.rw.Unlock()
		ch.cache[lr.ip] = newHolder
	}()
	metrics := ch.awaitingMetrics[lr.ip]
	if metrics != nil {
		delete(ch.awaitingMetrics, lr.ip)
		ch.statsMetricItemsQueued -= uint64(len(metrics))
		ch.statsMetricHostsQueued--
		go ch.updateAndDispatchMetrics(ctx, lr.instance, metrics...)
	}
	events := ch.awaitingEvents[lr.ip]
	if events != nil {
		delete(ch.awaitingEvents, lr.ip)
		ch.statsEventItemsQueued -= uint64(len(events))
		ch.statsEventHostsQueued--
		go ch.updateAndDispatchEvents(ctx, lr.instance, events...)
	}
}

func (ch *CloudHandler) handleMetric(ctx context.Context, m *gostatsd.Metric) {
	holder, ok := ch.cache[m.SourceIP]
	if ok {
		// While metric was in the queue the cache was primed. Use the value.
		holder.updateAccess()
		ch.statsCacheLateHit++
		go ch.updateAndDispatchMetrics(ctx, holder.instance, m)
	} else {
		// Still nothing in the cache.
		queue := ch.awaitingMetrics[m.SourceIP]
		ch.awaitingMetrics[m.SourceIP] = append(queue, m)
		if len(queue) == 0 {
			// This is the first metric in the queue
			ch.toLookupIPs = append(ch.toLookupIPs, m.SourceIP)
			ch.statsMetricHostsQueued++
		}
		ch.statsMetricItemsQueued++
		ch.statsCacheMiss++
	}
}

func (ch *CloudHandler) updateAndDispatchMetrics(ctx context.Context, instance *gostatsd.Instance, metrics ...*gostatsd.Metric) {
	for _, m := range metrics {
		updateInplace(&m.Tags, &m.Hostname, instance)
		if err := ch.metrics.DispatchMetric(ctx, m); err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return
			}
			log.Warnf("Failed to dispatch metric: %v", err)
		}
	}
}

func (ch *CloudHandler) handleEvent(ctx context.Context, e *gostatsd.Event) {
	holder, ok := ch.cache[e.SourceIP]
	if ok {
		// While event was in the queue the cache was primed. Use the value.
		holder.updateAccess()
		ch.statsCacheLateHit++
		go ch.updateAndDispatchEvents(ctx, holder.instance, e)
	} else {
		// Still nothing in the cache.
		queue := ch.awaitingEvents[e.SourceIP]
		ch.awaitingEvents[e.SourceIP] = append(queue, e)
		if len(queue) == 0 {
			// This is the first event in the queue
			ch.toLookupIPs = append(ch.toLookupIPs, e.SourceIP)
			ch.statsEventHostsQueued++
		}
		ch.statsEventItemsQueued++
		ch.statsCacheMiss++
	}
}

func (ch *CloudHandler) updateAndDispatchEvents(ctx context.Context, instance *gostatsd.Instance, events ...*gostatsd.Event) {
	var dispatched int
	defer func() {
		ch.wg.Add(-dispatched)
	}()
	for _, e := range events {
		updateInplace(&e.Tags, &e.Hostname, instance)
		dispatched++
		if err := ch.events.DispatchEvent(ctx, e); err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return
			}
			log.Warnf("Failed to dispatch event: %v", err)
		}
	}
}

func (ch *CloudHandler) updateTagsAndHostname(ip gostatsd.IP, tags *gostatsd.Tags, hostname *string) bool {
	instance, ok := ch.getInstance(ip)
	if ok {
		updateInplace(tags, hostname, instance)
	}
	return ok
}

func (ch *CloudHandler) getInstance(ip gostatsd.IP) (*gostatsd.Instance, bool) {
	if ip == gostatsd.UnknownIP {
		return nil, true
	}
	ch.rw.RLock()
	holder, ok := ch.cache[ip]
	ch.rw.RUnlock()
	if ok {
		holder.updateAccess()
		return holder.instance, true
	}
	return nil, false
}

func updateInplace(tags *gostatsd.Tags, hostname *string, instance *gostatsd.Instance) {
	if instance != nil { // It was a positive cache hit (successful lookup cache, not failed lookup cache)
		// Update hostname inplace
		*hostname = instance.ID
		// Update tag list inplace
		*tags = append(*tags, instance.Tags...)
	}
}
