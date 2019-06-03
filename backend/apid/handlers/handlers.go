package handlers

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// Handlers represents the HTTP handlers for CRUD operations on resources
type Handlers struct {
	Resource corev2.Resource
	Store    store.ResourceStore
}
