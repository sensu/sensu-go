package mockstore

import (
	"context"
	"errors"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/sensu/sensu-go/types"
)

var mu sync.Mutex
var rings = make(map[string]types.Ring)

// GetRing ...
func (s *MockStore) GetRing(parts ...string) types.Ring {
	mu.Lock()
	defer mu.Unlock()
	name := path.Join(parts...)
	if rings[name] == nil {
		rings[name] = &Ring{data: make(map[int64]string)}
	}
	return rings[name]
}

// Ring ...
type Ring struct {
	data map[int64]string
	mu   sync.Mutex
	err  error
}

// Add ...
func (r *Ring) Add(ctx context.Context, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	for _, v := range r.data {
		if v == value {
			return nil
		}
	}
	r.data[time.Now().UnixNano()] = value
	return nil
}

// Remove ...
func (r *Ring) Remove(ctx context.Context, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return r.err
	}
	for k, v := range r.data {
		if v == value {
			delete(r.data, k)
		}
	}
	return nil
}

// Next ...
func (r *Ring) Next(context.Context) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return "", r.err
	}
	if len(r.data) == 0 {
		return "", errors.New("empty ring")
	}
	keys := make([]int64, 0, len(r.data))
	for k := range r.data {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	value := r.data[keys[0]]
	delete(r.data, keys[0])
	r.data[time.Now().UnixNano()] = value
	return value, nil
}

// SetError will cause Ring's methods to return err.
func (r *Ring) SetError(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.err = err
}

// Peek ...
func (r *Ring) Peek(context.Context) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return "", r.err
	}
	if len(r.data) == 0 {
		return "", errors.New("empty ring")
	}
	keys := make([]int64, 0, len(r.data))
	for k := range r.data {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return r.data[keys[0]], nil
}
