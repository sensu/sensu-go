package v2

import (
	"github.com/prometheus/client_golang/prometheus"
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

// WatchActionType indicates what type of change was made to an object in the store.
type WatchActionType int

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
