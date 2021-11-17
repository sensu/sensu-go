package kvc

import (
	"context"
	"strings"
	"time"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/util/retry"
)

const (
	// EtcdRoot is the root of all sensu storage.
	EtcdRoot = "/sensu.io"

	// EtcdInitialDelay is 100 ms.
	EtcdInitialDelay = time.Millisecond * 100

	// NamespacesPathPrefix is the path to namespaces in etcds
	NamespacesPathPrefix = "namespaces"
)

// Backoff delivers a pre-configured backoff object, suitable for use in making
// etcd requests.
func Backoff(ctx context.Context) *retry.ExponentialBackoff {
	return &retry.ExponentialBackoff{
		Ctx:                  ctx,
		InitialDelayInterval: EtcdInitialDelay,
		MaxDelayInterval:     10 * time.Second,
		Multiplier:           5,
	}
}

// RetryRequest will return whether or not to try a request again based on the
// error given to it, and the number of times the request has been tried.
//
// If RetryRequest gets "etcdserver: too many requests", then it will return
// (false, nil). Otherwise, it will return (true, err).
func RetryRequest(n int, err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if err == context.Canceled {
		return true, err
	}
	if err == context.DeadlineExceeded {
		return true, err
	}
	// using string comparison here because it's too difficult to tell
	// what kind of error the client is actually delivering
	if strings.Contains(err.Error(), "etcdserver: too many requests") {
		if n > 3 {
			// don't log the first few retries, to avoid unnecessary log spam
			logger.WithError(err).WithField("retry", n).Error("retrying")
		}
		return false, nil
	}
	return true, &store.ErrInternal{Message: err.Error()}
}
