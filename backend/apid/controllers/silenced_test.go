package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHttpApiSilencedGet(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &SilencedController{
		Store: store,
	}

	silenced := []*types.Silenced{
		types.FixtureSilenced("check1"),
		types.FixtureSilenced("check2"),
	}
	store.On("GetSilencedEntries", mock.Anything).Return(silenced, nil)
	req := newRequest("GET", "/silenced", nil)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	returnedSilencedEntries := []*types.Silenced{}
	err := json.Unmarshal(body, &returnedSilencedEntries)

	assert.NoError(t, err)
	assert.EqualValues(t, silenced, returnedSilencedEntries)
}

func TestHttpApiSilencedGetError(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &SilencedController{
		Store: store,
	}

	var nilSilenced []*types.Silenced
	store.On("GetSilencedEntries", mock.Anything).Return(nilSilenced, errors.New("error"))
	req := newRequest("GET", "/silenced", nil)
	res := processRequest(c, req)

	body := res.Body.Bytes()

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "error\n", string(body))
}

func TestHttpApiSilencedGetUnauthorized(t *testing.T) {
	controller := SilencedController{}

	req := newRequest("GET", "/silenced", nil)
	req = requestWithNoAccess(req)

	res := processRequest(&controller, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestHttpApiSilencedPost(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &SilencedController{
		Store: store,
	}

	silenced := types.FixtureSilenced("check1")
	silenced.Subscription = "test-subscription"
	silenced.CheckName = "test-check"
	silenced.ID = silenced.Subscription + ":" + silenced.CheckName

	store.On("UpdateSilencedEntry", mock.Anything, silenced).Return(nil)

	encoded, _ := json.Marshal(silenced)

	req := newRequest("POST", "/silenced", bytes.NewBuffer(encoded))
	res := processRequest(c, req)

	assert.Equal(t, http.StatusCreated, res.Code)
}

// Expect an error to be returned
func TestHttpApiSilencedPostMissingCheckNameSubscription(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &SilencedController{
		Store: store,
	}

	silenced := types.FixtureSilenced("check1")
	silenced.Subscription = ""
	silenced.CheckName = ""

	store.On("UpdateSilencedEntry", mock.Anything, silenced).Return(errors.New("must provide a subscription or check name"))

	encoded, _ := json.Marshal(silenced)

	req := newRequest("POST", "/silenced", bytes.NewBuffer(encoded))
	res := processRequest(c, req)

	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestHttpApiSilencedClear(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &SilencedController{
		Store: store,
	}

	silenced := types.FixtureSilenced("check1")
	silenced.ID = "test-subcription:test-check"
	silenced.Subscription = "test-subscription"
	silenced.CheckName = "test-check"

	encoded, _ := json.Marshal(silenced)

	store.On("DeleteSilencedEntry", mock.Anything, silenced.ID).Return(nil)

	req := newRequest("POST", "/silenced/clear", bytes.NewBuffer(encoded))
	res := processRequest(c, req)

	assert.Equal(t, http.StatusNoContent, res.Code)

}

func TestHttpApiSilencedDeleteUnauthorized(t *testing.T) {
	controller := SilencedController{}

	silenced := types.FixtureSilenced("check1")
	silenced.ID = "test-subcription:test-check"
	silenced.Subscription = "test-subscription"
	silenced.CheckName = "test-check"

	encoded, _ := json.Marshal(silenced)
	req := newRequest("POST", "/silenced/clear", bytes.NewBuffer(encoded))
	req = requestWithNoAccess(req)

	res := processRequest(&controller, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestHttpApiSilencedGetByCheckName(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &SilencedController{
		Store: store,
	}

	silenced := []*types.Silenced{
		types.FixtureSilenced("check1"),
		types.FixtureSilenced("check2"),
	}
	store.On("GetSilencedEntriesByCheckName", mock.Anything).Return(silenced, nil)
	req := newRequest("GET", "/silenced/checks/check1", nil)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	returnedSilencedEntries := []*types.Silenced{}
	err := json.Unmarshal(body, &returnedSilencedEntries)

	assert.NoError(t, err)
	assert.EqualValues(t, silenced, returnedSilencedEntries)
}

func TestHttpApiSilencedGetBySubscription(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &SilencedController{
		Store: store,
	}

	check1 := types.FixtureSilenced("check1")
	check1.Subscription = "subscription1"
	check2 := types.FixtureSilenced("check2")
	check1.Subscription = "subscription1"
	silenced := []*types.Silenced{
		check1,
		check2,
	}

	store.On("GetSilencedEntriesBySubscription", mock.Anything).Return(silenced, nil)
	req := newRequest("GET", "/silenced/subscriptions/subscription1", nil)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	returnedSilencedEntries := []*types.Silenced{}
	err := json.Unmarshal(body, &returnedSilencedEntries)

	assert.NoError(t, err)
	assert.EqualValues(t, silenced, returnedSilencedEntries)
}
