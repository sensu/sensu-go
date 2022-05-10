package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"path"
	"sync"
	"time"

	"github.com/lib/pq"
)

// Bus is a postgresql message bus. It demultiplexes asynchronous notifications
// based on channel name to subscribers.
//
// There can only be one subscriber per unique (namespace, name) subscription.
type Bus struct {
	listener      Listener
	subscriptions map[string]*subscriptionSet
	mu            sync.Mutex
}

// ListenChannelName turns a namespace and a name into a channel name suitable
// for use with postgres asynchronous notification. As postgres channels must
// be 63 bytes long or fewer, if the path.Join of namespace and name would
// result in a name longer than 63 bytes, the SHA256 sum of the joined name
// is used instead.
func ListenChannelName(namespace, subName string) string {
	const maxChannelNameLength = 63
	name := path.Join(namespace, subName)
	if len(name) > maxChannelNameLength {
		sum := sha256.Sum256([]byte(name))
		name = string(base64.URLEncoding.EncodeToString(sum[:]))
	}
	return name
}

// Listener is the abstraction of a pq.Listener.
type Listener interface {
	Listen(string) error
	Unlisten(string) error
	UnlistenAll() error
	Close() error
	NotificationChannel() <-chan *pq.Notification
}

// NewBus creates a new Bus. It starts a goroutine that listens for
// notifications and demuxes them to any subscribers that are subscribed.
func NewBus(ctx context.Context, listener Listener) *Bus {
	bus := &Bus{
		listener:      listener,
		subscriptions: make(map[string]*subscriptionSet),
	}
	go bus.demuxNotifications(ctx)
	return bus
}

func (b *Bus) demuxNotifications(ctx context.Context) {
	notifications := b.listener.NotificationChannel()
	for {
		select {
		case <-ctx.Done():
			return
		case notification, ok := <-notifications:
			if !ok {
				return
			}
			b.demux(notification)
		}
	}
}

func (b *Bus) demux(notification *pq.Notification) {
	if notification == nil {
		return
	}
	channels := b.getNotificationChans(notification.Channel)
	for _, ch := range channels {
		select {
		case ch <- notification:
		case <-time.After(time.Minute):
			logger.Error("postgres notification not handled after 1 minute!")
		}
	}
}

func (b *Bus) getNotificationChans(chanName string) []chan *pq.Notification {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.subscriptions[chanName].Channels()
}

func (b *Bus) newNotificationChan(ctx context.Context, namespace, name string) (chan *pq.Notification, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := ListenChannelName(namespace, name)
	_, ok := b.subscriptions[key]
	if !ok {
		if err := b.listener.Listen(key); err != nil {
			return nil, err
		}
		b.subscriptions[key] = &subscriptionSet{}
	}
	ch := make(chan *pq.Notification, 8)
	b.subscriptions[key].Append(subscription{Notification: ch, Ctx: ctx})
	return ch, nil
}

// Subscribe starts listening for notifications for the (namespace, name)
// combination provided. Any number of goroutines can listen concurrently to a
// given namespace and name.
func (b *Bus) Subscribe(ctx context.Context, namespace, name string) (<-chan *pq.Notification, error) {
	return b.newNotificationChan(ctx, namespace, name)
}

type subscription struct {
	Notification chan *pq.Notification
	Ctx          context.Context
}

type subscriptionSet struct {
	mu            sync.Mutex
	subscriptions []subscription
}

func (s *subscriptionSet) Append(sub subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscriptions = append(s.subscriptions, sub)
	go s.removeWhenDone(sub)
}

func (s *subscriptionSet) Channels() []chan *pq.Notification {
	s.mu.Lock()
	defer s.mu.Unlock()
	channels := make([]chan *pq.Notification, len(s.subscriptions))
	for i, sub := range s.subscriptions {
		channels[i] = sub.Notification
	}
	return channels
}

func (s *subscriptionSet) removeWhenDone(sub subscription) {
	<-sub.Ctx.Done()
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := -1
	for i, ss := range s.subscriptions {
		if sub == ss {
			idx = i
			break
		}
	}
	if idx >= 0 {
		s.subscriptions = append(s.subscriptions[:idx], s.subscriptions[idx+1:]...)
	}
}
