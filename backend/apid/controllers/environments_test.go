package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteEnv(t *testing.T) {
	testCases := []struct {
		org            string
		env            string
		storeResponse  error
		expectedStatus int
	}{
		{"default", "default", fmt.Errorf("error"), http.StatusInternalServerError},
		{"default", "default", nil, http.StatusAccepted},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf(
			"when the store respond %s, the API should respond with %d",
			tc.storeResponse,
			tc.expectedStatus,
		)
		t.Run(testName, func(t *testing.T) {
			store := &mockstore.MockStore{}
			controller := &EnvironmentsController{Store: store}

			store.On("DeleteEnvironment", mock.Anything, tc.org, tc.org).Return(tc.storeResponse)
			req, _ := http.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("/rbac/organizations/%s/environments/%s", tc.org, tc.env),
				nil,
			)
			res := processRequest(controller, req)

			assert.Equal(t, tc.expectedStatus, res.Code)
		})
	}
}

func TestGetEnv(t *testing.T) {
	var nilEnv *types.Environment
	type storeResponse struct {
		env *types.Environment
		err error
	}
	testCases := []struct {
		org            string
		env            string
		storeResponse  storeResponse
		expectedStatus int
	}{
		{"missing", "default", storeResponse{nilEnv, fmt.Errorf("error")}, http.StatusInternalServerError},
		{"default", "missing", storeResponse{nilEnv, nil}, http.StatusNotFound},
		{"default", "default", storeResponse{types.FixtureEnvironment("default"), nil}, http.StatusOK},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf(
			"when the store respond %s, the API should respond with %d",
			tc.storeResponse,
			tc.expectedStatus,
		)
		t.Run(testName, func(t *testing.T) {
			store := &mockstore.MockStore{}
			controller := &EnvironmentsController{Store: store}

			store.On("GetEnvironment", mock.Anything, tc.org, tc.env).Return(tc.storeResponse.env, tc.storeResponse.err)

			req, _ := http.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/rbac/organizations/%s/environments/%s", tc.org, tc.env),
				nil,
			)

			res := processRequest(controller, req)
			assert.Equal(t, tc.expectedStatus, res.Code)

			body := res.Body.Bytes()

			if tc.expectedStatus == http.StatusOK {
				result := &types.Environment{}
				err := json.Unmarshal(body, &result)
				assert.NoError(t, err)
				assert.EqualValues(t, tc.storeResponse.env.Name, result.Name)
			}
		})
	}
}

func TestGetEnvs(t *testing.T) {
	testCases := []struct {
		org            string
		storeResponse  error
		expectedStatus int
	}{
		{"missing", fmt.Errorf("error"), http.StatusInternalServerError},
		{"default", nil, http.StatusOK},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf(
			"when the store respond %s, the API should respond with %d",
			tc.storeResponse,
			tc.expectedStatus,
		)
		t.Run(testName, func(t *testing.T) {
			store := &mockstore.MockStore{}
			controller := &EnvironmentsController{Store: store}

			env1 := types.FixtureEnvironment("foo")
			env2 := types.FixtureEnvironment("bar")
			envs := []*types.Environment{env1, env2}

			store.On("GetEnvironments", mock.Anything, tc.org).Return(envs, tc.storeResponse)

			req, _ := http.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/rbac/organizations/%s/environments", tc.org),
				nil,
			)

			res := processRequest(controller, req)
			assert.Equal(t, tc.expectedStatus, res.Code)

			body := res.Body.Bytes()

			if tc.expectedStatus == http.StatusOK {
				result := []*types.Environment{}
				err := json.Unmarshal(body, &result)
				assert.NoError(t, err)
				assert.EqualValues(t, envs, result)
			} else {
				assert.Equal(t, "error\n", string(body))
			}
		})
	}
}

func TestUpdateEnv(t *testing.T) {
	testCases := []struct {
		storeResponse  error
		expectedStatus int
	}{
		{fmt.Errorf("error"), http.StatusInternalServerError},
		{nil, http.StatusCreated},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf(
			"when the store respond %s, the API should respond with %d",
			tc.storeResponse,
			tc.expectedStatus,
		)
		t.Run(testName, func(t *testing.T) {
			store := &mockstore.MockStore{}
			controller := &EnvironmentsController{Store: store}

			store.On(
				"UpdateEnvironment",
				mock.Anything,
				mock.AnythingOfType("string"),
				mock.AnythingOfType("*types.Environment"),
			).Return(tc.storeResponse)

			env := types.FixtureEnvironment("foo")
			envBytes, _ := json.Marshal(env)

			req, _ := http.NewRequest(
				http.MethodPost,
				"/rbac/organizations/default/environments",
				bytes.NewBuffer(envBytes),
			)
			res := processRequest(controller, req)

			assert.Equal(t, tc.expectedStatus, res.Code)
		})
	}
}
