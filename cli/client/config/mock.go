package config

import (
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"

	"github.com/stretchr/testify/mock"
)

// MockConfig is a configuration used for CLI testing. When using the
// MockConfig in unit tests, stub out the behavior you wish to test against
// by assigning the appropriate function to the appropriate Func field. If you
// have forgotten to stub a particular function, the program will panic.
type MockConfig struct {
	mock.Mock
}

// APIUrl mocks the API URL config
func (m *MockConfig) APIUrl() string {
	args := m.Called()
	return args.String(0)
}

// Format mocks the format config
func (m *MockConfig) Format() string {
	args := m.Called()
	return args.String(0)
}

// InsecureSkipTLSVerify mocks the insecure skip TLS verify config
func (m *MockConfig) InsecureSkipTLSVerify() bool {
	args := m.Called()
	return args.Bool(0)
}

// Namespace mocks the namespace config
func (m *MockConfig) Namespace() string {
	args := m.Called()
	return args.String(0)
}

// Timeout mocks the timeout config
func (m *MockConfig) Timeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

// TrustedCAFile mocks the trusted CA file config
func (m *MockConfig) TrustedCAFile() string {
	args := m.Called()
	return args.String(0)
}

// SaveAPIUrl mocks saving the API URL
func (m *MockConfig) SaveAPIUrl(url string) error {
	args := m.Called(url)
	return args.Error(0)
}

// SaveFormat mocks saving the format
func (m *MockConfig) SaveFormat(format string) error {
	args := m.Called(format)
	return args.Error(0)
}

// SaveInsecureSkipTLSVerify ...
func (m *MockConfig) SaveInsecureSkipTLSVerify(verify bool) error {
	args := m.Called(verify)
	return args.Error(0)
}

// SaveNamespace mocks saving the namespace
func (m *MockConfig) SaveNamespace(namespace string) error {
	args := m.Called(namespace)
	return args.Error(0)
}

// SaveTimeout mocks saving the timeout
func (m *MockConfig) SaveTimeout(timeout time.Duration) error {
	args := m.Called(timeout)
	return args.Error(0)
}

// SaveTokens mocks saving the tokens
func (m *MockConfig) SaveTokens(tokens *corev2.Tokens) error {
	args := m.Called(tokens)
	return args.Error(0)
}

// SaveTrustedCAFile ...
func (m *MockConfig) SaveTrustedCAFile(file string) error {
	args := m.Called(file)
	return args.Error(0)
}

// Tokens mocks the tokens config
func (m *MockConfig) Tokens() *corev2.Tokens {
	args := m.Called()
	return args.Get(0).(*corev2.Tokens)
}
