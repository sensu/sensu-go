package controllers

import (
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestHttpApiEventsGetEmpty(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &EventsController{
		Store: store,
	}

	var events []*types.Event
	store.On("GetEvents").Return(events, nil)
	req, _ := http.NewRequest("GET", "/events", nil)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.String()
	assert.Equal(t, "[]", body)
}
