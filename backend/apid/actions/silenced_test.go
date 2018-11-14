package actions

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/dgrijalva/jwt-go"

	"github.com/sensu/sensu-go/testing/mockstore"
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
}

func TestSilencedQuery(t *testing.T) {
	defaultCtx := context.Background()

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
			results, err := actions.Query(tc.ctx, tc.params["subscription"], tc.params["check"])

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestSilencedFind(t *testing.T) {
	defaultCtx := context.Background()

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
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewSilencedController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetSilencedEntryByName", tc.ctx, mock.Anything, mock.Anything).
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

func TestSilencedCreateOrReplace(t *testing.T) {
	defaultCtx := context.Background()

	claims := &types.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}
	jwtCtx := context.WithValue(context.Background(), types.ClaimsKey, claims)

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
			name:        "Already Exists",
			ctx:         defaultCtx,
			argument:    types.FixtureSilenced("*:silence1"),
			fetchResult: types.FixtureSilenced("*:silence1"),
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
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badSilence,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Creator",
			ctx:             jwtCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			expectedErr:     false,
			expectedCreator: "foo",
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewSilencedController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetSilencedEntryByName", mock.Anything, mock.Anything).
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
			err := actions.CreateOrReplace(tc.ctx, *tc.argument)

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
func TestSilencedCreate(t *testing.T) {
	defaultCtx := context.Background()

	claims := &types.Claims{StandardClaims: jwt.StandardClaims{Subject: "foo"}}
	jwtCtx := context.WithValue(context.Background(), types.ClaimsKey, claims)

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
		expectedId      string
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureSilenced("*:silence1"),
			expectedErr: false,
			expectedId:  "*:silence1",
		},
		{
			name:            "Already Exists",
			ctx:             defaultCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			fetchResult:     types.FixtureSilenced("*:silence1"),
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
			expectedId:      "*:silence1",
		},
		{
			name:            "Store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
			expectedId:      "*:silence1",
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
			expectedId:      "*:silence1",
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badSilence,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
			expectedId:      "*:silence1",
		},
		{
			name:            "Creator",
			ctx:             jwtCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			expectedErr:     false,
			expectedCreator: "foo",
			expectedId:      "*:silence1",
		},
		{
			name:            "Other Id",
			ctx:             jwtCtx,
			argument:        types.FixtureSilenced("unix:*"),
			expectedErr:     false,
			expectedCreator: "foo",
			expectedId:      "unix:*",
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewSilencedController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetSilencedEntryByName", mock.Anything, mock.Anything).
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
						assert.Equal(tc.expectedId, entry.Name)
					}
				})

			// Exec Query
			err := actions.Create(tc.ctx, tc.argument)

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
	defaultCtx := context.Background()

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
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewSilencedController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetSilencedEntryByName", mock.Anything, mock.Anything).
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
	defaultCtx := context.Background()

	testCases := []struct {
		name            string
		ctx             context.Context
		id              string
		deleteErr       error
		fetchErr        error
		fetchResult     *types.Silenced
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Deleted",
			ctx:         defaultCtx,
			id:          "i-424242:*",
			fetchResult: types.FixtureSilenced("i-424242:*"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			id:              "missing:*",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			id:              "i-424242:*",
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store Err on Delete",
			ctx:             defaultCtx,
			id:              "i-424242:*",
			deleteErr:       errors.New("dunno"),
			fetchResult:     types.FixtureSilenced("i-424242:*"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewSilencedController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetSilencedEntryByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("DeleteSilencedEntryByName", mock.Anything, mock.Anything).
				Return(tc.deleteErr).Once()

			// Exec Query
			err := actions.Destroy(tc.ctx, tc.id)

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
