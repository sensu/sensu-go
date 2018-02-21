package actions

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewSilencedController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewSilencedController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.Store)
	assert.NotNil(actions.Policy)
}

func TestSilencedQuery(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithPerms(types.RuleTypeSilenced, types.RulePermRead),
	)

	testCases := []struct {
		name         string
		ctx          context.Context
		params       QueryParams
		storeErr     error
		storeRecords []*types.Silenced
		expectedLen  int
		expectedErr  error
	}{
		{
			name:         "No Silenced Entries",
			ctx:          defaultCtx,
			storeRecords: []*types.Silenced{},
			expectedLen:  0,
			storeErr:     nil,
			expectedErr:  nil,
		},
		{
			name: "With Silenced Entries",
			ctx:  defaultCtx,
			storeRecords: []*types.Silenced{
				types.FixtureSilenced("*:silence1"),
				types.FixtureSilenced("*:silence2"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:   "Filter By Subscription",
			ctx:    defaultCtx,
			params: QueryParams{"subscription": "test"},
			storeRecords: []*types.Silenced{
				types.FixtureSilenced("*:silence1"),
				types.FixtureSilenced("*:silence2"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:   "Filter By Check",
			ctx:    defaultCtx,
			params: QueryParams{"check": "test"},
			storeRecords: []*types.Silenced{
				types.FixtureSilenced("*:silence1"),
				types.FixtureSilenced("*:silence2"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "With Only Create Access",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeSilenced, types.RulePermCreate),
			)),
			storeRecords: []*types.Silenced{
				types.FixtureSilenced("*:silence1"),
				types.FixtureSilenced("*:silence2"),
			},
			expectedLen: 0,
			storeErr:    nil,
		},
		{
			name:         "Store Failure",
			ctx:          defaultCtx,
			storeRecords: nil,
			expectedLen:  0,
			storeErr:     errors.New(""),
			expectedErr:  NewError(InternalErr, errors.New("")),
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewSilencedController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.On("GetSilencedEntriesBySubscription", tc.ctx).Return(tc.storeRecords, tc.storeErr).Once()
			store.On("GetSilencedEntriesByCheckName", tc.ctx).Return(tc.storeRecords, tc.storeErr).Once()
			store.On("GetSilencedEntries", tc.ctx).Return(tc.storeRecords, tc.storeErr).Once()

			// Exec Query
			results, err := actions.Query(tc.ctx, tc.params)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestSilencedFind(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithPerms(types.RuleTypeSilenced, types.RulePermRead),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		record          *types.Silenced
		argument        string
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "No argument given",
			ctx:             defaultCtx,
			argument:        "",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name:            "Found",
			ctx:             defaultCtx,
			record:          types.FixtureSilenced("*:silence1"),
			argument:        "silence1",
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "Not Found",
			ctx:             defaultCtx,
			record:          nil,
			argument:        "missing",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name: "No Read Permission",
			ctx: testutil.NewContext(testutil.ContextWithPerms(
				types.RuleTypeSilenced,
				types.RulePermCreate,
			)),
			record:          types.FixtureSilenced("*:silence1"),
			argument:        "silence1",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewSilencedController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetSilencedEntryByID", tc.ctx, mock.Anything, mock.Anything).
				Return(tc.record, nil).Once()

			// Exec Query
			result, err := actions.Find(tc.ctx, tc.argument)

			inferErr, ok := err.(Error)
			if ok {
				assert.Equal(tc.expectedErrCode, inferErr.Code)
			} else {
				assert.NoError(err)
			}
			assert.Equal(tc.expected, result != nil, "expects Find() to return a record")
		})
	}
}

func TestSilencedCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithPerms(types.RuleTypeSilenced, types.RulePermCreate),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithPerms(types.RuleTypeSilenced, types.RulePermRead),
	)
	actorCtx := testutil.NewContext(
		testutil.ContextWithActor("actorID", types.FixtureRuleWithPerms(types.RuleTypeSilenced, types.RulePermCreate)),
	)

	badSilence := types.FixtureSilenced("*:silence1")
	badSilence.Check = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Silenced
		fetchResult     *types.Silenced
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
		expectedCreator string
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureSilenced("*:silence1"),
			expectedErr: false,
		},
		{
			name:            "Already Exists",
			ctx:             defaultCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			fetchResult:     types.FixtureSilenced("*:silence1"),
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
		},
		{
			name:            "Store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badSilence,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Creator",
			ctx:             actorCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			expectedErr:     false,
			expectedCreator: "actorID",
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewSilencedController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetSilencedEntryByID", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateSilencedEntry", mock.Anything, mock.Anything).
				Return(tc.createErr).
				Run(func(args mock.Arguments) {
					if tc.expectedCreator != "" {
						bytes, _ := json.Marshal(args[1])
						entry := types.Silenced{}
						_ = json.Unmarshal(bytes, &entry)

						assert.Equal(tc.expectedCreator, entry.Creator)
					}
				})

			// Exec Query
			err := actions.Create(tc.ctx, *tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Given was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestSilencedUpdate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithPerms(types.RuleTypeSilenced, types.RulePermUpdate),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithPerms(types.RuleTypeSilenced, types.RulePermRead),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Silenced
		fetchResult     *types.Silenced
		fetchErr        error
		updateErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Updated",
			ctx:         defaultCtx,
			argument:    types.FixtureSilenced("*:silence1"),
			fetchResult: types.FixtureSilenced("*:silence1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Update",
			ctx:             defaultCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			fetchResult:     types.FixtureSilenced("*:silence1"),
			updateErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			fetchResult:     types.FixtureSilenced("*:silence1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			fetchResult:     types.FixtureSilenced("*:silence1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewSilencedController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetSilencedEntryByID", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateSilencedEntry", mock.Anything, mock.Anything).
				Return(tc.updateErr)

			// Exec Query
			err := actions.Update(tc.ctx, *tc.argument)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Given was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestSilencedDestroy(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithPerms(types.RuleTypeSilenced, types.RulePermDelete),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithPerms(types.RuleTypeSilenced, types.RulePermCreate),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		params          QueryParams
		fetchResult     *types.Silenced
		deleteErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Deleted",
			ctx:         defaultCtx,
			params:      QueryParams{"id": "silence1"},
			fetchResult: types.FixtureSilenced("*:silence1"),
			expectedErr: false,
		},
		{
			name:        "Subscription Params",
			ctx:         defaultCtx,
			params:      QueryParams{"subscription": "test"},
			fetchResult: types.FixtureSilenced("*:silence1"),
			expectedErr: false,
		},
		{
			name:        "Check Param",
			ctx:         defaultCtx,
			params:      QueryParams{"check": "test"},
			fetchResult: types.FixtureSilenced("*:silence1"),
			expectedErr: false,
		},
		{
			name:            "Store Err on Delete",
			ctx:             defaultCtx,
			params:          QueryParams{"id": "silence1"},
			fetchResult:     types.FixtureSilenced("*:silence1"),
			deleteErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No Permission",
			ctx:             wrongPermsCtx,
			params:          QueryParams{"id": "silence1"},
			fetchResult:     types.FixtureSilenced("*:silence1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewSilencedController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("DeleteSilencedEntryByID", mock.Anything, mock.Anything).
				Return(tc.deleteErr).Once()
			store.
				On("DeleteSilencedEntriesByCheckName", mock.Anything, mock.Anything).
				Return(tc.deleteErr).Once()
			store.
				On("DeleteSilencedEntriesBySubscription", mock.Anything, mock.Anything).
				Return(tc.deleteErr).Once()

			// Exec Query
			err := actions.Destroy(tc.ctx, tc.params)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Given was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}
