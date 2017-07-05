package controllers

import (
	"net/http"
	"testing"

	"github.com/gorilla/context"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/stretchr/testify/assert"
)

func TestOrganization(t *testing.T) {
	// Empty query parameter
	r, _ := http.NewRequest("GET", "/", nil)
	org := organization(r)
	assert.Equal(t, defaultOrganization, org)

	// Org in context
	contextOrg := "bar"
	r, _ = http.NewRequest("GET", "/", nil)
	context.Set(r, middlewares.OrganizationKey, contextOrg)

	org = organization(r)
	assert.Equal(t, contextOrg, org)
}
