package testing

import (
	"net/http"

	"github.com/sensu/sensu-go/cli/client"

	corev2 "github.com/sensu/core/v2"
)

// CreateSilenced for use with mock lib
func (c *MockClient) CreateSilenced(silenced *corev2.Silenced) error {
	args := c.Called(silenced)
	return args.Error(0)
}

// UpdateSilenced for use with mock lib
func (c *MockClient) UpdateSilenced(silenced *corev2.Silenced) error {
	args := c.Called(silenced)
	return args.Error(0)
}

// DeleteSilenced for use with mock lib
func (c *MockClient) DeleteSilenced(namespace, name string) error {
	args := c.Called(namespace, name)
	return args.Error(0)
}

// FetchSilenced for use with mock lib
func (c *MockClient) FetchSilenced(id string) (*corev2.Silenced, error) {
	args := c.Called(id)
	return args.Get(0).(*corev2.Silenced), args.Error(1)
}

// ListSilenceds for use with mock lib
func (c *MockClient) ListSilenceds(namespace, sub, check string, options *client.ListOptions, header *http.Header) ([]corev2.Silenced, error) {
	args := c.Called(namespace, sub, check, options, header)
	return args.Get(0).([]corev2.Silenced), args.Error(1)
}
