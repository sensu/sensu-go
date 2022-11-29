package actions

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	jwt "github.com/golang-jwt/jwt/v4"
	corev2 "github.com/sensu/core/v2"
	coreJWT "github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
		sv2 := new(mockstore.V2MockStore)
		sv2.On("GetSilencesStore").Return(store)
		actions := NewSilencedController(sv2)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.On("GetSilencesBySubscription", tc.ctx, mock.Anything, mock.Anything).Return(tc.storeRecords, tc.storeErr).Once()
			store.On("GetSilencesByCheck", tc.ctx, mock.Anything, mock.Anything).Return(tc.storeRecords, tc.storeErr).Once()
			store.On("GetSilences", tc.ctx, mock.Anything).Return(tc.storeRecords, tc.storeErr).Once()

			// Exec Query
			results, err := actions.List(tc.ctx, tc.params["subscription"], tc.params["check"])

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
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
		sv2 := new(mockstore.V2MockStore)
		sv2.On("GetSilencesStore").Return(store)
		actions := NewSilencedController(sv2)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetSilenceByName", mock.Anything, mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateSilence", mock.Anything, mock.Anything).
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
			err := actions.CreateOrReplace(tc.ctx, tc.argument)

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
		expectedID      string
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureSilenced("*:silence1"),
			expectedErr: false,
			expectedID:  "*:silence1",
		},
		{
			name:            "Already Exists",
			ctx:             defaultCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			fetchResult:     types.FixtureSilenced("*:silence1"),
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
			expectedID:      "*:silence1",
		},
		{
			name:            "Store Err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
			expectedID:      "*:silence1",
		},
		{
			name:            "Store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
			expectedID:      "*:silence1",
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badSilence,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
			expectedID:      "*:silence1",
		},
		{
			name:            "Creator",
			ctx:             jwtCtx,
			argument:        types.FixtureSilenced("*:silence1"),
			expectedErr:     false,
			expectedCreator: "foo",
			expectedID:      "*:silence1",
		},
		{
			name:            "Other Id",
			ctx:             jwtCtx,
			argument:        types.FixtureSilenced("unix:*"),
			expectedErr:     false,
			expectedCreator: "foo",
			expectedID:      "unix:*",
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		sv2 := new(mockstore.V2MockStore)
		sv2.On("GetSilencesStore").Return(store)
		actions := NewSilencedController(sv2)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetSilenceByName", mock.Anything, mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateSilence", mock.Anything, mock.Anything, mock.Anything).
				Return(tc.createErr).
				Run(func(args mock.Arguments) {
					if tc.expectedCreator != "" {
						bytes, _ := json.Marshal(args[1])
						entry := types.Silenced{}
						_ = json.Unmarshal(bytes, &entry)

						assert.Equal(tc.expectedCreator, entry.Creator)
						assert.Equal(tc.expectedID, entry.Name)
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

func TestSilencedCreatedBy(t *testing.T) {
	claims, err := coreJWT.NewClaims(&corev2.User{Username: "admin"})
	assert.NoError(t, err)
	ctx := context.WithValue(context.Background(), corev2.ClaimsKey, claims)
	silenced := corev2.FixtureSilenced("silenced1:*")

	store := &mockstore.MockStore{}
	sv2 := new(mockstore.V2MockStore)
	sv2.On("GetSilencesStore").Return(store)
	actions := NewSilencedController(sv2)

	var s *corev2.Silenced
	store.On("UpdateSilence", mock.Anything, mock.Anything).Return(nil)
	store.On("GetSilenceByName", mock.Anything, mock.Anything, mock.Anything).Return(s, nil)

	err = actions.Create(ctx, silenced)
	assert.NoError(t, err)
	assert.Equal(t, "admin", silenced.CreatedBy)

	err = actions.CreateOrReplace(ctx, silenced)
	assert.NoError(t, err)
	assert.Equal(t, "admin", silenced.CreatedBy)
}
