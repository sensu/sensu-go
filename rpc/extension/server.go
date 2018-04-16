package extension

import (
	"context"
	"net"

	"github.com/sensu/sensu-go/rpc"
	"google.golang.org/grpc"
)

var _ rpc.ExtensionServer = &Server{}

// Server is a Sensu extension.
//
// Create new extensions that will run on a local port with NewServer, and
// then supply any of HandlerFunc, MutatorFunc or FilterFunc via the Handler
// Mutator
type Server struct {
	ext    Interface
	ln     net.Listener
	server *grpc.Server
}

// NewServer creates a new server that will run on the provided listener.
func NewServer(extension Interface, ln net.Listener) *Server {
	return &Server{
		ln:     ln,
		server: grpc.NewServer(),
		ext:    extension,
	}
}

// Start starts the extension service.
func (s *Server) Start() error {
	rpc.RegisterExtensionServer(s.server, s)
	go func() {
		// TODO: add error handling once gRPC errors are a little less
		// noisy.
		_ = s.server.Serve(s.ln)
	}()
	return nil
}

// HandleEvent is called by gRPC.
func (s *Server) HandleEvent(ctx context.Context, req *rpc.HandleEventRequest) (*rpc.HandleEventResponse, error) {
	err := s.ext.HandleEvent(req.Event, req.MutatedEvent)
	if err == ErrNotImplemented {
		return nil, err
	}
	resp := rpc.HandleEventResponse{}
	if err != nil {
		resp.Error = err.Error()
	}
	return &resp, nil
}

// MutateEvent is called by gRPC.
func (s *Server) MutateEvent(ctx context.Context, req *rpc.MutateEventRequest) (*rpc.MutateEventResponse, error) {
	mutated, err := s.ext.MutateEvent(req.Event)
	if err == ErrNotImplemented {
		return nil, err
	}
	resp := rpc.MutateEventResponse{
		MutatedEvent: mutated,
	}
	if err != nil {
		resp.Error = err.Error()
	}
	return &resp, nil
}

// FilterEvent is called by gRPC.
func (s *Server) FilterEvent(ctx context.Context, req *rpc.FilterEventRequest) (*rpc.FilterEventResponse, error) {
	filtered, err := s.ext.FilterEvent(req.Event)
	if err == ErrNotImplemented {
		return nil, err
	}
	resp := rpc.FilterEventResponse{
		Filtered: filtered,
	}
	if err != nil {
		resp.Error = err.Error()
	}
	return &resp, nil
}

// Stop stops the extension. If the extension's listener was pass in to the
// extension  with NewServerWithListener, then the listener will not be
// closed by Stop.
func (s *Server) Stop() error {
	s.server.GracefulStop()
	return s.ln.Close()
}
