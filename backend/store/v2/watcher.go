package v2

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	corev3 "github.com/sensu/core/v3"
	apitools "github.com/sensu/sensu-api-tools"
)

const (
	// WatchUnknown indicates that we received an unknown watch even tytpe
	// from etcd.
	WatchUnknown WatchActionType = iota
	// WatchCreate indicates that an object was created.
	WatchCreate
	// WatchUpdate indicates that an object was updated.
	WatchUpdate
	// WatchDelete indicates that an object was deleted.
	WatchDelete
	// WatchError indicates that an error was encountered
	WatchError
)

// WatcherOf is designed to work with interfaces that are implemented by
// multiple types of resources. It will discover which resource types implement
// T by scanning the type registry. For each concrete type that implements T,
// WatcherOf will start a goroutine and watch for it. All watch events will be
// multiplexed into a single channel.
//
// TODO: support things other than configuration store resources, like entities
func WatcherOf[T corev3.Resource](ctx context.Context, store Interface) Watcher {
	resourceTypes := apitools.FindTypesOf[T]()
	result := make(chan []WatchEvent, 1)
	for _, rType := range resourceTypes {
		req := NewResourceRequestFromResource(rType)
		watcher := store.GetConfigStore().Watch(ctx, req)
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case events := <-watcher:
					result <- events
				}
			}
		}()
	}
	return multiWatcher{
		events: result,
	}
}

type multiWatcher struct {
	events chan []WatchEvent
}

func (m multiWatcher) Result() <-chan []WatchEvent {
	return m.events
}

// WatchActionType indicates what type of change was made to an object in the store.
type WatchActionType int

func ReadEventValue[R Resource[T], T any](w WatchEvent) (R, error) {
	if w.Err != nil {
		return nil, w.Err
	}
	if w.Value == nil {
		return nil, nil
	}
	val := new(T)
	if err := w.Value.UnwrapInto(val); err != nil {
		return nil, err
	}
	return val, nil
}

// WatchEvent represents an event of a watched resource
type WatchEvent struct {
	Type          WatchActionType
	Key           ResourceRequest
	Value         Wrapper
	PreviousValue Wrapper
	Revision      int64
	Err           error
}

// Watcher represents a generic watcher
type Watcher interface {
	Result() <-chan []WatchEvent
}

func (t WatchActionType) String() string {
	var s string
	switch t {
	case WatchUnknown:
		s = "Unknown"
	case WatchCreate:
		s = "Create"
	case WatchDelete:
		s = "Delete"
	case WatchUpdate:
		s = "Update"
	case WatchError:
		s = "Error"
	}
	return s
}

const (
	WatchEventsCounterVec = "sensu_go_watch_events"

	WatchEventsLabelStatus       = "status"
	WatchEventsLabelResourceType = "resource"
	WatchEventsLabelNamespace    = "namespace"

	WatchEventsStatusHandled = "handled"
	WatchEventsStatusDropped = "dropped"

	WatcherProvider     = "provider"
	WatcherProviderPG   = "postgres"
	WatcherProviderEtcd = "etcd"
)

var (
	WatchEventsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: WatchEventsCounterVec,
			Help: "The total number of store watch notifications",
		},
		[]string{WatchEventsLabelStatus, WatchEventsLabelResourceType, WatchEventsLabelNamespace, WatcherProvider},
	)
)

func init() {
	if err := prometheus.Register(WatchEventsProcessed); err != nil {
		panic(err)
	}
}
