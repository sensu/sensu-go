package v2

import corev3 "github.com/sensu/sensu-go/api/core/v3"

type Action uint

const (
	Create Action = iota
	Update
	Delete
	WatchError
)

type WatchEvent struct {
	Action   Action
	Resource corev3.Resource
}

type Watcher interface {
	Watch(ResourceRequest) (<-chan WatchEvent, error)
}
