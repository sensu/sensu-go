package nodes

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

var (
	errNoNodes = errors.New("no nodes available")
)

type redisNodeTracker struct {
	client    *redis.Client
	namespace string
	nodeid    string
	nodes     nodeList

	updateInterval time.Duration
	expiryInterval time.Duration

	rw sync.RWMutex // protects the nodes list, not the individual members in the list

	// for testing
	now func() time.Time
}

// NewRedisNodeTracker returns a NodeTracker based on a Redis backend
func NewRedisNodeTracker(redisAddr, namespace, nodeid string, updateInterval, expiryInterval time.Duration) NodeTracker {
	options := &redis.Options{
		Addr: redisAddr,
		DB:   0,
	}

	return &redisNodeTracker{
		client:    redis.NewClient(options),
		namespace: namespace,
		nodeid:    nodeid,
		nodes:     make(nodeList, 0, 10),

		updateInterval: updateInterval,
		expiryInterval: expiryInterval,

		now: time.Now,
	}
}

// Run will track nodes via Redis PubSub until the context is closed.
func (rnt *redisNodeTracker) Run(ctx context.Context) {
	pubsub := rnt.client.Subscribe(rnt.namespace)
	defer pubsub.Close()

	ticker := time.NewTicker(rnt.updateInterval)
	defer ticker.Stop()

	psChan := pubsub.Channel() // Closed when pubsub is Closed

	rnt.emitPresence()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rnt.expireNodes()
			if rnt.nodeid != "" {
				err := rnt.emitPresence()
				if err != nil {
					logrus.Warningf("Failed to check in to cluster")
				}
			}
		case msg := <-psChan:
			// TODO: Validate msg
			rnt.updateNode(msg.Payload)
		}
	}
}

// List returns a list of all the nodes currently tracked.  Intended for admin
// interface, not general usage.  It is more expensive than Select.  Thread safe.
func (rnt *redisNodeTracker) List() []string {
	// Does not talk to Redis
	rnt.rw.RLock()
	defer rnt.rw.RUnlock()
	nodes := make([]string, len(rnt.nodes))
	for idx, node := range rnt.nodes {
		nodes[idx] = node.nodeid
	}
	return nodes
}

// Select will use the provided key to pick a node and return it.  An error will
// be returned if no nodes are available.  Thread safe.
func (rnt *redisNodeTracker) Select(key uint64) (string, error) {
	// Does not talk to Redis
	rnt.rw.RLock()
	defer rnt.rw.RUnlock()
	if len(rnt.nodes) == 0 {
		return "", errNoNodes
	}
	return rnt.nodes[key%uint64(len(rnt.nodes))].nodeid, nil
}

// expireNodes will expire nodes which have not updated recently enough. Thread safe,
// but takes the write lock.
func (rnt *redisNodeTracker) expireNodes() {
	// Does not talk to Redis
	now := rnt.now()
	rnt.rw.Lock()
	defer rnt.rw.Unlock()

	// https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
	nodes := rnt.nodes[:0]
	for _, node := range rnt.nodes {
		if now.Before(node.expiry) {
			// 	Keep it
			nodes = append(nodes, node)
		}
	}
	rnt.nodes = nodes
}

// updateNode will attempt to update the expiry on an existing node, if it
// doesn't exist, it will be added under the write lock.
func (rnt *redisNodeTracker) updateNode(nodeid string) {
	// Does not talk to Redis
	if rnt.tryUpdateExistingNode(nodeid) {
		return
	}

	node := &node{
		nodeid: nodeid,
		expiry: rnt.now().Add(5 * time.Second),
	}

	rnt.rw.Lock()
	defer rnt.rw.Unlock()
	rnt.nodes = append(rnt.nodes, node)
	sort.Sort(rnt.nodes)
}

// tryUpdateExistingNode will update a nodes expiry time in place if it exists
// and returns true if it succeeds.  Prevents taking the write lock.
func (rnt *redisNodeTracker) tryUpdateExistingNode(nodeid string) bool {
	// Does not talk to redis
	rnt.rw.RLock()
	defer rnt.rw.RUnlock()
	for _, node := range rnt.nodes {
		if node.nodeid == nodeid {
			node.expiry = rnt.now().Add(5 * time.Second)
			return true
		}
	}
	return false
}

// emitPresence will announce the presence of this node to the PubSub endpoint
func (rnt *redisNodeTracker) emitPresence() error {
	// Talks to redis
	cmd := rnt.client.Publish(rnt.namespace, rnt.nodeid)
	return cmd.Err()
}
