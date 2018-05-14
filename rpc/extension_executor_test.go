package rpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

type mockClient struct {
	mock.Mock
}

func (m *mockClient) FilterEvent(ctx context.Context, req *FilterEventRequest, opts ...grpc.CallOption) (*FilterEventResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*FilterEventResponse), args.Error(1)
}

func (m *mockClient) MutateEvent(ctx context.Context, req *MutateEventRequest, opts ...grpc.CallOption) (*MutateEventResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*MutateEventResponse), args.Error(1)
}

func (m *mockClient) HandleEvent(ctx context.Context, req *HandleEventRequest, opts ...grpc.CallOption) (*HandleEventResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*HandleEventResponse), args.Error(1)
}

func TestFilterEvent(t *testing.T) {
	tests := []struct {
		resp     *FilterEventResponse
		rpcErr   error
		filtered bool
		err      bool
	}{
		{
			resp: &FilterEventResponse{
				Filtered: false,
				Error:    "",
			},
			rpcErr:   nil,
			filtered: false,
			err:      false,
		},
		{
			resp: &FilterEventResponse{
				Filtered: true,
				Error:    "",
			},
			rpcErr:   nil,
			filtered: true,
			err:      false,
		},
		{
			resp: &FilterEventResponse{
				Filtered: false,
				Error:    "error",
			},
			rpcErr:   nil,
			filtered: false,
			err:      true,
		},
		{
			resp: &FilterEventResponse{
				Filtered: true,
				Error:    "",
			},
			rpcErr:   errors.New("an error"),
			filtered: false,
			err:      true,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			client := &mockClient{}
			client.On("FilterEvent", mock.Anything, mock.Anything).Return(test.resp, test.rpcErr)
			executor := &GRPCExtensionExecutor{client: client}
			filtered, err := executor.FilterEvent(types.FixtureEvent("foo", "bar"))
			if test.err && err == nil {
				t.Fatal("expected non-nil error")
			}
			if !test.err && err != nil {
				t.Fatal(err)
			}
			if got, want := filtered, test.filtered; got != want {
				t.Errorf("bad filtered: got %v, want %v", got, want)
			}
		})
	}
}

func TestMutateEvent(t *testing.T) {
	tests := []struct {
		resp    *MutateEventResponse
		rpcErr  error
		mutated []byte
		err     bool
	}{
		{
			resp: &MutateEventResponse{
				MutatedEvent: nil,
				Error:        "",
			},
			rpcErr:  nil,
			mutated: nil,
			err:     false,
		},
		{
			resp: &MutateEventResponse{
				MutatedEvent: []byte("^_^"),
				Error:        "",
			},
			rpcErr:  nil,
			mutated: []byte("^_^"),
			err:     false,
		},
		{
			resp: &MutateEventResponse{
				MutatedEvent: []byte("ya"),
				Error:        "error",
			},
			rpcErr:  nil,
			mutated: []byte("ya"),
			err:     true,
		},
		{
			resp: &MutateEventResponse{
				MutatedEvent: []byte("hi"),
				Error:        "",
			},
			rpcErr:  errors.New("an error"),
			mutated: nil,
			err:     true,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			client := &mockClient{}
			client.On("MutateEvent", mock.Anything, mock.Anything).Return(test.resp, test.rpcErr)
			executor := &GRPCExtensionExecutor{client: client}
			mutated, err := executor.MutateEvent(types.FixtureEvent("foo", "bar"))
			if test.err && err == nil {
				t.Fatal("expected non-nil error")
			}
			if !test.err && err != nil {
				t.Fatal(err)
			}
			if got, want := mutated, test.mutated; !bytes.Equal(got, want) {
				t.Errorf("bad mutated: got %v, want %v", got, want)
			}
		})
	}
}

func TestHandleEvent(t *testing.T) {
	tests := []struct {
		resp   *HandleEventResponse
		rpcErr error
		err    bool
	}{
		{
			resp: &HandleEventResponse{
				Error:  "",
				Output: "foo",
			},
			rpcErr: nil,
			err:    false,
		},
		{
			resp: &HandleEventResponse{
				Error:  "",
				Output: "foo",
			},
			rpcErr: nil,
			err:    false,
		},
		{
			resp: &HandleEventResponse{
				Error:  "error",
				Output: "error",
			},
			rpcErr: nil,
			err:    true,
		},
		{
			resp: &HandleEventResponse{
				Error:  "",
				Output: "error",
			},
			rpcErr: errors.New("an error"),
			err:    true,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			client := &mockClient{}
			client.On("HandleEvent", mock.Anything, mock.Anything, mock.Anything).Return(test.resp, test.rpcErr)
			executor := &GRPCExtensionExecutor{client: client}
			handlerResp, err := executor.HandleEvent(types.FixtureEvent("foo", "bar"), nil)
			if test.err && err == nil {
				t.Fatal("expected non-nil error")
			}
			if !test.err && err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, test.resp.Output, handlerResp.Output)
			assert.Equal(t, test.resp.Error, handlerResp.Error)
		})
	}
}
