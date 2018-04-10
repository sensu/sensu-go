package rpc

import (
	"context"
	"errors"
	"sync"

	"github.com/sensu/sensu-go/types"
	"google.golang.org/grpc"
)

type clientConnCache struct {
	sync.Mutex
	clients map[string]*grpc.ClientConn
}

// Get gets a ClientConn from the cache. ClientConns are safe for concurrent
// use according to the gRPC issue tracker.
func (c *clientConnCache) Get(service string) (*grpc.ClientConn, error) {
	c.Lock()
	defer c.Unlock()
	var err error
	conn, ok := c.clients[service]
	if !ok {
		conn, err = grpc.Dial(service)
		if err != nil {
			return nil, err
		}
		c.clients[service] = conn
	}
	return conn, err
}

var connCache = clientConnCache{
	clients: make(map[string]*grpc.ClientConn),
}

// ExtensionExecutor executes extensions.
type ExtensionExecutor interface {
	// FilterEvent filters an event.
	FilterEvent(*types.Event) (bool, error)

	// MutateEvent mutates an event.
	MutateEvent(*types.Event) ([]byte, error)

	// HandleEvent handles an event. It is passed both the original event
	// and the mutated event, if the event was mutated.
	HandleEvent(*types.Event, []byte) error
}

// GRPCExtensionExecutor executes extension rpc methods.
type GRPCExtensionExecutor struct {
	extension *types.Extension
	client    ExtensionClient
}

// NewGRPCExtensionExecutor creates a new GRPCExtensionExecutor.
func NewGRPCExtensionExecutor(ext *types.Extension) (ExtensionExecutor, error) {
	conn, err := connCache.Get(ext.URL)
	if err != nil {
		return nil, err
	}
	return &GRPCExtensionExecutor{
		extension: ext,
		client:    NewExtensionClient(conn),
	}, nil
}

// FilterEvent filters an event.
func (e *GRPCExtensionExecutor) FilterEvent(evt *types.Event) (bool, error) {
	resp, err := e.client.FilterEvent(context.Background(), &FilterEventRequest{Event: evt})
	if err != nil {
		return false, err
	}
	if resp.Error != "" {
		err = errors.New(resp.Error)
	}
	return resp.Filtered, err
}

// MutateEvent mutates an event.
func (e *GRPCExtensionExecutor) MutateEvent(evt *types.Event) ([]byte, error) {
	resp, err := e.client.MutateEvent(context.Background(), &MutateEventRequest{Event: evt})
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		err = errors.New(resp.Error)
	}
	return resp.MutatedEvent, err
}

// HandleEvent handles an event.
func (e *GRPCExtensionExecutor) HandleEvent(evt *types.Event, mutatedEvt []byte) error {
	req := &HandleEventRequest{Event: evt, MutatedEvent: mutatedEvt}
	resp, err := e.client.HandleEvent(context.Background(), req)
	if err != nil {
		return err
	}
	if resp.Error != "" {
		err = errors.New(resp.Error)
	}
	return err
}
