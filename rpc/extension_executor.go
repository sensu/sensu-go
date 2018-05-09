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
func (e *GRPCExtensionExecutor) HandleEvent(evt *types.Event, mutatedEvt []byte) (HandleEventResponse, error) {
	req := &HandleEventRequest{Event: evt, MutatedEvent: mutatedEvt}
	resp, err := e.client.HandleEvent(context.Background(), req)
	if err != nil && resp == nil {
		return HandleEventResponse{}, err
	}
	if err == nil && resp.Error != "" {
		err = errors.New(resp.Error)
	}
	return *resp, err
}
