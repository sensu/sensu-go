package statsd

import (
	"context"
	"time"

	"github.com/atlassian/gostatsd"

	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

const (
	maxLookupIPs  = 32
	batchDuration = 10 * time.Millisecond
)

type lookupDispatcher struct {
	limiter       *rate.Limiter
	cloud         gostatsd.CloudProvider // Cloud provider interface
	toLookup      <-chan gostatsd.IP
	lookupResults chan<- *lookupResult
}

func (ld *lookupDispatcher) run(ctx context.Context) {
	ips := make([]gostatsd.IP, 0, maxLookupIPs)
	var c <-chan time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case ip := <-ld.toLookup:
			ips = append(ips, ip)
			if len(ips) >= maxLookupIPs {
				break // enough ips, exit select
			}
			if c == nil {
				c = time.After(batchDuration)
			}
			continue // have some more time to collect ips
		case <-c: // time to do the lookup
		}
		c = nil

		if err := ld.limiter.Wait(ctx); err != nil {
			if err != context.Canceled && err != context.DeadlineExceeded {
				// This could be an error caused by context signaling done.
				// Or something nasty but it is very unlikely.
				log.Warnf("Error from limiter: %v", err)
			}
			return
		}
		ld.doLookup(ctx, ips)
		for i := range ips { // cleanup pointers for GC
			ips[i] = gostatsd.UnknownIP
		}
		ips = ips[:0] // reset slice
	}
}

func (ld *lookupDispatcher) doLookup(ctx context.Context, ips []gostatsd.IP) {
	// instances may contain partial result even if err != nil
	instances, err := ld.cloud.Instance(ctx, ips...)
	for _, ip := range ips {
		instance := instances[ip]
		var res *lookupResult
		if instance != nil {
			res = &lookupResult{
				ip:       ip,
				instance: instance,
			}

		} else {
			res = &lookupResult{
				err: err,
				ip:  ip,
			}
		}
		select {
		case <-ctx.Done():
			return
		case ld.lookupResults <- res:
		}
	}
}
