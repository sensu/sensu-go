package mocktransport

import (
	"net/http"

	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

// MockTransport ...
type MockTransport struct {
	mock.Mock
}

// Close ...
func (m *MockTransport) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Closed ...
func (m *MockTransport) Closed() bool {
	args := m.Called()
	return args.Bool(0)
}

// Receive ...
func (m *MockTransport) Receive() (*transport.Message, error) {
	args := m.Called()
	return args.Get(0).(*transport.Message), args.Error(1)
}

// Reconnect ...
func (m *MockTransport) Reconnect(wsServerURL string, tlsOpts *types.TLSOptions, requestHeader http.Header) error {
	args := m.Called(wsServerURL, tlsOpts, requestHeader)
	return args.Error(0)
}

// Send ...
func (m *MockTransport) Send(message *transport.Message) error {
	args := m.Called(message)
	return args.Error(0)
}
