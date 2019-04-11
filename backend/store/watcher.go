package store

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
	Type   WatchActionType
	Key    string
	Object []byte
	Err    error
}

// Watcher represents a generic watcher
type Watcher interface {
	Result() <-chan WatchEvent
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
