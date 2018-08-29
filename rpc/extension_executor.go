package rpc

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sensu/sensu-go/types"
	"google.golang.org/grpc"
)

const ExtensionTimeout = 5 * time.Second

type clientConnCache struct {
	sync.Mutex
	clients map[string]*grpc.ClientConn
}

// Get gets a ClientConn from the cache. ClientConns are safe for concurrent
// use according to the gRPC issue tracker.
func (c *clientConnCache) Get(service string) (*grpc.ClientConn, error) {
	// TODO: implement gRPC connection caching
	// The cache needs to monitor its connection flock and make sure that any
	// connections that get into a bad state are purged from the cache.
	//
	// gRPC has an experimental API for this: https://godoc.org/google.golang.org/grpc#ClientConn.WaitForStateChange
	// And also provides metrics: https://godoc.org/google.golang.org/grpc#Server.ChannelzMetric
	return grpc.Dial(service, grpc.WithInsecure())
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
	HandleEvent(*types.Event, []byte) (HandleEventResponse, error)

	// Close closes the underlying TCP connection of the executor
	Close() error
}

// GRPCExtensionExecutor executes extension rpc methods.
type GRPCExtensionExecutor struct {
	extension *types.Extension
	client    ExtensionClient
	conn      *grpc.ClientConn
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
		conn:      conn,
	}, nil
}

// Close closes the extension executor. It will not be usable after Close is called.
func (e *GRPCExtensionExecutor) Close() error {
	return e.conn.Close()
}

// FilterEvent filters an event.
func (e *GRPCExtensionExecutor) FilterEvent(evt *types.Event) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ExtensionTimeout)
	defer cancel()
	resp, err := e.client.FilterEvent(ctx, &FilterEventRequest{Event: evt})
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
	ctx, cancel := context.WithTimeout(context.Background(), ExtensionTimeout)
	defer cancel()
	resp, err := e.client.MutateEvent(ctx, &MutateEventRequest{Event: evt})
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		err = errors.New(resp.Error)
	}
	return resp.MutatedEvent, err
}

// HandleEvent handles an event.
func (e *GRPCExtensionExecutor) HandleEvent(evt *types.Event, mutatedEvt []byte) (HandleEventResponse, error) {
	req := &HandleEventRequest{Event: evt, MutatedEvent: mutatedEvt}
	ctx, cancel := context.WithTimeout(context.Background(), ExtensionTimeout)
	defer cancel()
	resp, err := e.client.HandleEvent(ctx, req)
	if err != nil && resp == nil {
		return HandleEventResponse{}, err
	}
	if err == nil && resp.Error != "" {
		err = errors.New(resp.Error)
	}
	return *resp, err
}
