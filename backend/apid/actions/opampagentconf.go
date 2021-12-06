package actions

import (
	"context"
	"sync"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
)

type OpampAgentConfController struct {
	mu          sync.Mutex
	store       string
	contentType string
}

func (c *OpampAgentConfController) Get(_ context.Context) (*corev3.OpampAgentConfig, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return &corev3.OpampAgentConfig{
		Body:        c.store,
		ContentType: c.contentType,
	}, nil
}

func (c *OpampAgentConfController) CreateOrUpdate(_ context.Context, config *corev3.OpampAgentConfig) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = config.Body
	c.contentType = config.ContentType
	return nil
}
