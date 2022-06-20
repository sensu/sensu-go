package etcdstore

import (
	"context"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	watchOpts = []clientv3.OpOption{
		clientv3.WithCreatedNotify(),
		clientv3.WithPrevKV(),
	}
)

func (s *Store) Watch(ctx context.Context, req storev2.ResourceRequest) <-chan []storev2.WatchEvent {
	logger.Infof("watching %s: %s/%s", req.StoreName, req.Namespace, req.Name)
	outbox := make(chan []storev2.WatchEvent, 1)
	go s.watchLoop(ctx, req, outbox)
	return outbox
}

func (s *Store) watchLoop(ctx context.Context, req storev2.ResourceRequest, outbox chan []storev2.WatchEvent) {
	defer logger.Infof("stopped watching %s: %s/%s", req.StoreName, req.Namespace, req.Name)
	defer close(outbox)
	ctx = clientv3.WithRequireLeader(ctx)
	key := StoreKey(req)
	prefix := req.Name == ""
	if prefix && !strings.HasSuffix(key, "/") {
		key += "/"
	}
	for {
		watcher := s.newWatchChan(ctx, key, prefix)
		then := time.Now()
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-watcher:
				if !ok {
					goto RESTART
				}
				if err := e.Err(); err != nil && !e.Canceled {
					logger.Error(err)
				}
				if len(e.Events) == 0 {
					continue
				}
				events := v2WatchEvents(req, e)
				var status string
				select {
				case outbox <- events:
					status = storev2.WatchEventsStatusHandled
				default:
					status = storev2.WatchEventsStatusDropped
				}
				storev2.WatchEventsProcessed.WithLabelValues(
					status,
					req.StoreName,
					req.Namespace,
					storev2.WatcherProviderEtcd,
				).Add(float64(len(events)))
			}
		}
	RESTART:
		logger.Infof("restarting watcher %s: %s/%s", req.StoreName, req.Namespace, req.Name)
		// don't permit the creation of more than 1 watcher/s per key
		since := time.Since(then)
		if since < time.Second {
			time.Sleep(time.Second - since)
		}
	}
}

func (s *Store) newWatchChan(ctx context.Context, key string, prefix bool) clientv3.WatchChan {
	opts := watchOpts
	if prefix {
		opts = append(opts, clientv3.WithPrefix())
	}
	return s.client.Watch(ctx, key, opts...)
}

func v2WatchEvents(req storev2.ResourceRequest, resp clientv3.WatchResponse) []storev2.WatchEvent {
	events := make([]storev2.WatchEvent, len(resp.Events))
	for i, event := range resp.Events {
		events[i] = toWatchEvent(req, event)
	}
	return events
}

func toWatchEvent(req storev2.ResourceRequest, event *clientv3.Event) storev2.WatchEvent {
	result := storev2.WatchEvent{Key: req}
	if event.Kv != nil {
		var wrapper wrap.Wrapper
		if err := proto.Unmarshal(event.Kv.Value, &wrapper); err != nil {
			result.Err = err
		}
		result.Value = &wrapper
	}
	if event.PrevKv != nil {
		var wrapper wrap.Wrapper
		if err := proto.Unmarshal(event.PrevKv.Value, &wrapper); err != nil {
			result.Err = err
		}
		result.PreviousValue = &wrapper
	}
	switch event.Type {
	case mvccpb.PUT:
		if event.PrevKv == nil {
			result.Type = storev2.WatchCreate
		} else {
			result.Type = storev2.WatchUpdate
		}
	case mvccpb.DELETE:
		result.Type = storev2.WatchDelete
	}
	return result
}
