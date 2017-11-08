package actions

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestNewMutatorController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	ctl := NewMutatorController(store)
	assert.NotNil(ctl)
	assert.Equal(store, ctl.Store)
	assert.NotNil(ctl.Policy)
}

func TestMutatorQuery(t *testing.T) {
	readCtx := testutil.NewContext(testutil.ContextWithRules(
		types.FixtureRuleWithPerms(types.RuleTypeMutator, types.RulePermRead)))

	tests := []struct {
		name        string
		ctx         context.Context
		mutators    []*types.Mutator
		params      QueryParams
		expectedLen int
		storeErr    error
		expectedErr error
	}{
		{
			name:        "No Params, No Mutators",
			ctx:         readCtx,
			mutators:    nil,
			params:      QueryParams{},
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "No Params With Mutators",
			ctx:  readCtx,
			mutators: []*types.Mutator{
				types.FixtureMutator("homer"),
				types.FixtureMutator("bart"),
			},
			params:      QueryParams{},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "No Params With Only Create Access",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeMutator, types.RulePermCreate),
			)),
			mutators: []*types.Mutator{
				types.FixtureMutator("lisa"),
				types.FixtureMutator("maggie"),
			},
			params:      QueryParams{},
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "Mutator Param",
			ctx:  readCtx,
			mutators: []*types.Mutator{
				types.FixtureMutator("mr. burns"),
			},
			params: QueryParams{
				"name": "mr. burns",
			},
			expectedLen: 1,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:     "Store Failure",
			ctx:      readCtx,
			mutators: nil,
			params: QueryParams{
				"name": "ralph",
			},
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, test := range tests {
		store := &mockstore.MockStore{}
		ctl := NewMutatorController(store)

		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.On("GetMutators", test.ctx).Return(test.mutators, test.storeErr)

			results, err := ctl.Query(test.ctx, test.params)

			assert.EqualValues(test.expectedErr, err)
			assert.Len(results, test.expectedLen)
		})
	}
}

func TestMutatorFind(t *testing.T) {
	readCtx := testutil.NewContext(testutil.ContextWithRules(
		types.FixtureRuleWithPerms(types.RuleTypeMutator, types.RulePermRead),
	))

	tests := []struct {
		name            string
		ctx             context.Context
		mutator         *types.Mutator
		params          QueryParams
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "No Params",
			ctx:             readCtx,
			params:          QueryParams{},
			expected:        false,
			expectedErrCode: InvalidArgument,
		},
		{
			name:    "Found",
			ctx:     readCtx,
			mutator: types.FixtureMutator("abe"),
			params: QueryParams{
				"name": "abe",
			},
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:    "Not Found",
			ctx:     readCtx,
			mutator: nil,
			params: QueryParams{
				"name": "fox mulder",
			},
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name: "No Read Permission",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeEvent, types.RulePermCreate),
			)),
			mutator: types.FixtureMutator("troy maclure"),
			params: QueryParams{
				"name": "troy maclure",
			},
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &mockstore.MockStore{}
			ctl := NewMutatorController(store)

			// Mock store methods
			store.On("GetMutatorByName", test.ctx, test.params["name"]).
				Return(test.mutator, nil)

			assert := assert.New(t)
			result, err := ctl.Find(test.ctx, test.params)
			if cerr, ok := err.(Error); ok {
				assert.Equal(test.expectedErrCode, cerr.Code)
			} else {
				assert.NoError(err)
			}
			assert.Equal(test.expected, result != nil, "expects Find() to return an event")
		})
	}
}

// func TestHttpApiMutatorsGet(t *testing.T) {
// 	store := &mockstore.MockStore{}
// 	controller := MutatorController{Store: store}
//
// 	mutators := []*types.Mutator{
// 		types.FixtureMutator("mutator1"),
// 		types.FixtureMutator("mutator2"),
// 	}
//
// 	store.On("GetMutators", mock.Anything).Return(mutators, nil)
// 	req := newRequest("GET", "/mutators", nil)
// 	res := processRequest(&controller, req)
//
// 	assert.Equal(t, http.StatusOK, res.Code)
//
// 	body := res.Body.Bytes()
//
// 	returnedMutators := []*types.Mutator{}
// 	err := json.Unmarshal(body, &returnedMutators)
//
// 	assert.NoError(t, err)
// 	assert.Equal(t, 2, len(returnedMutators))
// 	for i, mutator := range returnedMutators {
// 		assert.EqualValues(t, mutators[i], mutator)
// 	}
// }
//
// func TestHttpApiMutatorsGetUnauthorized(t *testing.T) {
// 	controller := MutatorController{}
//
// 	req := newRequest("GET", "/mutators", nil)
// 	req = requestWithNoAccess(req)
//
// 	res := processRequest(&controller, req)
// 	assert.Equal(t, http.StatusUnauthorized, res.Code)
// }
//
// func TestHttpApiMutatorGet(t *testing.T) {
// 	store := &mockstore.MockStore{}
//
// 	c := &MutatorController{
// 		Store: store,
// 	}
//
// 	var nilMutator *types.Mutator
// 	store.On("GetMutatorByName", mock.Anything, "somemutator").Return(nilMutator, nil)
// 	notFoundReq := newRequest("GET", "/mutators/somemutator", nil)
// 	notFoundRes := processRequest(c, notFoundReq)
//
// 	assert.Equal(t, http.StatusNotFound, notFoundRes.Code)
//
// 	mutatorName := "mutator1"
// 	mutator := types.FixtureMutator(mutatorName)
// 	store.On("GetMutatorByName", mock.Anything, mutatorName).Return(mutator, nil)
// 	foundReq := newRequest("GET", fmt.Sprintf("/mutators/%s", mutatorName), nil)
// 	foundRes := processRequest(c, foundReq)
//
// 	assert.Equal(t, http.StatusOK, foundRes.Code)
//
// 	body := foundRes.Body.Bytes()
//
// 	returnedMutator := &types.Mutator{}
// 	err := json.Unmarshal(body, &returnedMutator)
//
// 	assert.NoError(t, err)
// 	assert.EqualValues(t, mutator, returnedMutator)
// }
//
// func TestHttpApiMutatorGetUnauthorized(t *testing.T) {
// 	store := &mockstore.MockStore{}
// 	controller := MutatorController{Store: store}
//
// 	mutator := types.FixtureMutator("name")
// 	store.On("GetMutatorByName", mock.Anything, "name").Return(mutator, nil)
//
// 	req := newRequest("GET", "/mutators/name", nil)
// 	req = requestWithNoAccess(req)
//
// 	res := processRequest(&controller, req)
// 	assert.Equal(t, http.StatusUnauthorized, res.Code)
// }
//
// func TestHttpApiMutatorPut(t *testing.T) {
// 	store := &mockstore.MockStore{}
//
// 	c := &MutatorController{
// 		Store: store,
// 	}
//
// 	mutatorName := "mutator1"
// 	mutator := types.FixtureMutator(mutatorName)
//
// 	updatedMutatorJSON, _ := json.Marshal(mutator)
//
// 	store.On("UpdateMutator", mock.AnythingOfType("*types.Mutator")).Return(nil).Run(func(args mock.Arguments) {
// 		receivedMutator := args.Get(0).(*types.Mutator)
// 		assert.EqualValues(t, mutator, receivedMutator)
// 	})
// 	putReq := newRequest("PUT", fmt.Sprintf("/mutators/%s", mutatorName), bytes.NewBuffer(updatedMutatorJSON))
// 	putRes := processRequest(c, putReq)
//
// 	assert.Equal(t, http.StatusOK, putRes.Code)
// }
//
// func TestHttpApiMutatorPost(t *testing.T) {
// 	store := &mockstore.MockStore{}
//
// 	c := &MutatorController{
// 		Store: store,
// 	}
//
// 	mutatorName := "newmutator1"
// 	mutator := types.FixtureMutator(mutatorName)
//
// 	updatedMutatorJSON, _ := json.Marshal(mutator)
//
// 	store.On("UpdateMutator", mock.AnythingOfType("*types.Mutator")).Return(nil).Run(func(args mock.Arguments) {
// 		receivedMutator := args.Get(0).(*types.Mutator)
// 		assert.EqualValues(t, mutator, receivedMutator)
// 	})
// 	putReq := newRequest("POST", fmt.Sprintf("/mutators/%s", mutatorName), bytes.NewBuffer(updatedMutatorJSON))
// 	putRes := processRequest(c, putReq)
//
// 	assert.Equal(t, http.StatusOK, putRes.Code)
//
// 	unauthReq := newRequest("POST", "/mutators/"+mutatorName, bytes.NewBuffer(updatedMutatorJSON))
// 	unauthReq = requestWithNoAccess(unauthReq)
//
// 	unauthRes := processRequest(c, unauthReq)
// 	assert.Equal(t, http.StatusUnauthorized, unauthRes.Code)
// }
//
// func TestHttpApiMutatorDelete(t *testing.T) {
// 	store := &mockstore.MockStore{}
//
// 	c := &MutatorController{
// 		Store: store,
// 	}
//
// 	mutatorName := "mutator1"
// 	mutator := types.FixtureMutator(mutatorName)
// 	store.On("GetMutatorByName", mock.Anything, mutatorName).Return(mutator, nil)
// 	store.On("DeleteMutatorByName", mock.Anything, mutatorName).Return(nil)
// 	deleteReq := newRequest("DELETE", fmt.Sprintf("/mutators/%s", mutatorName), nil)
// 	deleteRes := processRequest(c, deleteReq)
//
// 	assert.Equal(t, http.StatusOK, deleteRes.Code)
// }
//
// func TestHttpApiMutatorDeleteUnauthorized(t *testing.T) {
// 	controller := MutatorController{}
//
// 	req := newRequest("DELETE", "/mutators/test", nil)
// 	req = requestWithNoAccess(req)
//
// 	res := processRequest(&controller, req)
// 	assert.Equal(t, http.StatusUnauthorized, res.Code)
// }
