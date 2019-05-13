package v2

import (
	"errors"
	fmt "fmt"
	"net/url"
	"sort"
	"strings"
)

const (
	// HandlerPipeType represents handlers that pipes event data // into arbitrary
	// commands via STDIN
	HandlerPipeType = "pipe"

	// HandlerSetType represents handlers that groups event handlers, making it
	// easy to manage groups of actions that should be executed for certain types
	// of events.
	HandlerSetType = "set"

	// HandlerTCPType represents handlers that send event data to a remote TCP
	// socket
	HandlerTCPType = "tcp"

	// HandlerUDPType represents handlers that send event data to a remote UDP
	// socket
	HandlerUDPType = "udp"

	// HandlerGRPCType is a special kind of handler that represents an extension
	HandlerGRPCType = "grpc"
)

// NewHandler creates a new Handler.
func NewHandler(meta ObjectMeta) *Handler {
	return &Handler{ObjectMeta: meta}
}

// Validate returns an error if the handler does not pass validation tests.
func (h *Handler) Validate() error {
	if err := ValidateName(h.Name); err != nil {
		return errors.New("handler name " + err.Error())
	}

	if err := h.validateType(); err != nil {
		return err
	}

	if h.Namespace == "" {
		return errors.New("namespace must be set")
	}

	return nil
}

func (h *Handler) validateType() error {
	if h.Type == "" {
		return errors.New("empty handler type")
	}

	switch h.Type {
	case "pipe", "set", "grpc":
		return nil
	case "tcp", "udp":
		return h.Socket.Validate()
	}

	return fmt.Errorf("unknown handler type: %s", h.Type)
}

// Validate returns an error if the handler socket does not pass validation tests.
func (s *HandlerSocket) Validate() error {
	if s == nil {
		return errors.New("tcp and udp handlers need a valid socket")
	}
	if len(s.Host) == 0 {
		return errors.New("socket host undefined")
	}
	if s.Port == 0 {
		return errors.New("socket port undefined")
	}
	return nil
}

//
// Sorting

type cmpHandler func(a, b *Handler) bool

// SortHandlersByPredicate is used to sort a given collection using a given predicate.
func SortHandlersByPredicate(hs []*Handler, fn cmpHandler) sort.Interface {
	return &handlerSorter{handlers: hs, byFn: fn}
}

// SortHandlersByName is used to sort a given collection of handlers by their names.
func SortHandlersByName(hs []*Handler, asc bool) sort.Interface {
	if asc {
		return SortHandlersByPredicate(hs, func(a, b *Handler) bool {
			return a.Name < b.Name
		})
	}

	return SortHandlersByPredicate(hs, func(a, b *Handler) bool {
		return a.Name > b.Name
	})
}

type handlerSorter struct {
	handlers []*Handler
	byFn     cmpHandler
}

// Len implements sort.Interface
func (s *handlerSorter) Len() int {
	return len(s.handlers)
}

// Swap implements sort.Interface
func (s *handlerSorter) Swap(i, j int) {
	s.handlers[i], s.handlers[j] = s.handlers[j], s.handlers[i]
}

// Less implements sort.Interface
func (s *handlerSorter) Less(i, j int) bool {
	return s.byFn(s.handlers[i], s.handlers[j])
}

// FixtureHandler returns a Handler fixture for testing.
func FixtureHandler(name string) *Handler {
	return &Handler{
		Type:       HandlerPipeType,
		Command:    "command",
		ObjectMeta: NewObjectMeta(name, "default"),
	}
}

// FixtureSocketHandler returns a Handler fixture for testing.
func FixtureSocketHandler(name string, proto string) *Handler {
	handler := FixtureHandler(name)
	handler.Type = proto
	handler.Socket = &HandlerSocket{
		Host: "127.0.0.1",
		Port: 3001,
	}
	return handler
}

// FixtureSetHandler returns a Handler fixture for testing.
func FixtureSetHandler(name string, handlers ...string) *Handler {
	handler := FixtureHandler(name)
	handler.Handlers = handlers
	return handler
}

// URIPath returns the path component of a Handler URI.
func (h *Handler) URIPath() string {
	return fmt.Sprintf("/api/core/v2/namespaces/%s/handlers/%s", url.PathEscape(h.Namespace), url.PathEscape(h.Name))
}

// HandlerFields returns a set of fields that represent that resource
func HandlerFields(r Resource) map[string]string {
	resource := r.(*Handler)
	return map[string]string{
		"handler.name":      resource.ObjectMeta.Name,
		"handler.namespace": resource.ObjectMeta.Namespace,
		"handler.filters":   strings.Join(resource.Filters, ","),
		"handler.handlers":  strings.Join(resource.Handlers, ","),
		"handler.mutator":   resource.Mutator,
		"handler.type":      resource.Type,
	}
}
