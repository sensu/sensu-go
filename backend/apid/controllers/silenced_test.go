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

	testCases := []struct {
		description        string
		silencedEntry      *types.Silenced
		expectedStatusCode int
	}{
		{
			"post silenced entry with checkname and subscription",
			&types.Silenced{
				Subscription: "test-subscription",
				CheckName:    "test-check",
				ID:           "test-subscription:test-check",
			},
			http.StatusCreated,
		},
		{
			"post silenced entry with no checkname",
			&types.Silenced{
				Subscription: "test-subscription",
				CheckName:    "",
				ID:           "test-subscription:*",
			},
			http.StatusCreated,
		},
		{
			"post silenced entry with no subscription",
			&types.Silenced{
				Subscription: "",
				CheckName:    "test-check",
				ID:           "*:test-check",
			},
			http.StatusCreated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			store.On("UpdateSilencedEntry", mock.Anything, tc.silencedEntry).Return(nil)

			encoded, _ := json.Marshal(tc.silencedEntry)

			req := newRequest("POST", "/silenced", bytes.NewBuffer(encoded))
			res := processRequest(c, req)

			assert.Equal(t, tc.expectedStatusCode, res.Code)
		})
	}
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

func TestHttpApiSilencedDelete(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &SilencedController{
		Store: store,
	}

	testCases := []struct {
		description        string
		route              string
		storeMethod        string
		id                 string
		expectedStatusCode int
	}{
		{
			"delete silenced entry with checkname and subscription",
			"/silenced/test-subscription:test-check",
			"DeleteSilencedEntryByID",
			"test-subscription:test-check",
			http.StatusNoContent,
		},
		{
			"delete silenced entry with no checkname",
			"/silenced/subscriptions/test-subscription",
			"DeleteSilencedEntriesBySubscription",
			"test-subscription",
			http.StatusNoContent,
		},
		{
			"delete silenced entry with no subscription",
			"/silenced/checks/test-check",
			"DeleteSilencedEntriesByCheckName",
			"test-check",
			http.StatusNoContent,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			store.On(tc.storeMethod, mock.Anything, tc.id).Return(nil)

			req := newRequest("DELETE", tc.route, nil)
			res := processRequest(c, req)

			assert.Equal(t, tc.expectedStatusCode, res.Code)
		})
	}
}

func TestHttpApiSilencedDeleteUnauthorized(t *testing.T) {
	controller := SilencedController{}

	req := newRequest("DELETE", "/silenced/test-subscription:test-check", nil)
	req = requestWithNoAccess(req)

	res := processRequest(&controller, req)
	assert.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestHttpApiSilencedGetByID(t *testing.T) {
	store := &mockstore.MockStore{}

	c := &SilencedController{
		Store: store,
	}
	silenced := []*types.Silenced{
		&types.Silenced{
			CheckName:    "check1",
			Subscription: "subscription1",
			ID:           "subscription1:check1",
		},
	}

	store.On("GetSilencedEntryByID", mock.Anything, "subscription1:check1").Return(silenced, nil)
	req := newRequest("GET", "/silenced/subscription1:check1", nil)
	res := processRequest(c, req)

	assert.Equal(t, http.StatusOK, res.Code)

	body := res.Body.Bytes()

	returnedSilencedEntries := []*types.Silenced{}
	err := json.Unmarshal(body, &returnedSilencedEntries)

	assert.NoError(t, err)
	assert.EqualValues(t, silenced, returnedSilencedEntries)
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
