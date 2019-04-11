package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewMutatorController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	ctl := NewMutatorController(store)
	assert.NotNil(ctl)
	assert.Equal(store, ctl.store)
}

func TestMutatorCreateOrReplace(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badMut := types.FixtureMutator("bad")
	badMut.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	tests := []struct {
		name            string
		ctx             context.Context
		argument        *types.Mutator
		fetchResult     *types.Mutator
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureMutator("sleepy"),
			expectedErr: false,
		},
		{
			name:        "Already Exists",
			ctx:         defaultCtx,
			argument:    types.FixtureMutator("sleepy"),
			fetchResult: types.FixtureMutator("sleepy"),
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badMut,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, test := range tests {
		store := &mockstore.MockStore{}
		ctl := NewMutatorController(store)

		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			store.On("GetMutatorByName", mock.Anything, mock.Anything).
				Return(test.fetchResult, test.fetchErr)

			store.On("UpdateMutator", mock.Anything, mock.Anything).Return(test.createErr)

			err := ctl.CreateOrReplace(test.ctx, *test.argument)

			if test.expectedErr {
				if cerr, ok := err.(Error); ok {
					assert.Equal(test.expectedErrCode, cerr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestMutatorCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badMut := types.FixtureMutator("bad")
	badMut.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	tests := []struct {
		name            string
		ctx             context.Context
		argument        *types.Mutator
		fetchResult     *types.Mutator
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureMutator("sleepy"),
			expectedErr: false,
		},
		{
			name:            "Already Exists",
			ctx:             defaultCtx,
			argument:        types.FixtureMutator("sleepy"),
			fetchResult:     types.FixtureMutator("sleepy"),
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
		},
		{
			name:            "store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureMutator("grumpy"),
			fetchErr:        errors.New("nein"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Validation Error",
			ctx:             defaultCtx,
			argument:        badMut,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, test := range tests {
		store := &mockstore.MockStore{}
		ctl := NewMutatorController(store)

		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			store.On("GetMutatorByName", mock.Anything, mock.Anything).
				Return(test.fetchResult, test.fetchErr)

			store.On("UpdateMutator", mock.Anything, mock.Anything).Return(test.createErr)

			err := ctl.Create(test.ctx, *test.argument)

			if test.expectedErr {
				if cerr, ok := err.(Error); ok {
					assert.Equal(test.expectedErrCode, cerr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Not of type 'Error'")
				}
			} else {
				assert.NoError(err)
			}
		})
	}
}

func TestMutatorDestroy(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		mutator         string
		fetchResult     *types.Mutator
		fetchErr        error
		deleteErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Deleted",
			ctx:         defaultCtx,
			mutator:     "mutator1",
			fetchResult: types.FixtureMutator("mutator1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			mutator:         "mutator1",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "store Err on Delete",
			ctx:             defaultCtx,
			mutator:         "mutator1",
			fetchResult:     types.FixtureMutator("mutator1"),
			deleteErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "store Err on Fetch",
			ctx:             defaultCtx,
			mutator:         "mutator1",
			fetchResult:     types.FixtureMutator("mutator1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewMutatorController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetMutatorByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("DeleteMutatorByName", mock.Anything, "mutator1").
				Return(tc.deleteErr)

			// Exec Query
			err := actions.Destroy(tc.ctx, tc.mutator)

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

func TestMutatorList(t *testing.T) {
	readCtx := context.Background()

	testCases := []struct {
		name        string
		ctx         context.Context
		records     []*types.Mutator
		storeErr    error
		expectedLen int
		expectedErr error
	}{
		{
			name:        "No Params, No Mutators",
			ctx:         readCtx,
			records:     nil,
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "No Params With Mutators",
			ctx:  readCtx,
			records: []*types.Mutator{
				types.FixtureMutator("homer"),
				types.FixtureMutator("bart"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "Mutator Param",
			ctx:  readCtx,
			records: []*types.Mutator{
				types.FixtureMutator("mr. burns"),
			},
			expectedLen: 1,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:        "store Failure",
			ctx:         readCtx,
			records:     nil,
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, tc := range testCases {
		s := &mockstore.MockStore{}
		actions := NewMutatorController(s)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			pred := &store.SelectionPredicate{}
			s.On("GetMutators", tc.ctx, pred).Return(tc.records, tc.storeErr)

			// Exec Query
			results, err := actions.List(tc.ctx, pred)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestMutatorFind(t *testing.T) {
	readCtx := context.Background()

	tests := []struct {
		name            string
		ctx             context.Context
		mutator         *types.Mutator
		argument        string
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "Found",
			ctx:             readCtx,
			mutator:         types.FixtureMutator("abe"),
			argument:        "abe",
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "Not Found",
			ctx:             readCtx,
			mutator:         nil,
			argument:        "fox mulder",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &mockstore.MockStore{}
			ctl := NewMutatorController(store)

			// Mock store methods
			store.
				On("GetMutatorByName", test.ctx, test.argument).
				Return(test.mutator, nil)

			assert := assert.New(t)
			result, err := ctl.Find(test.ctx, test.argument)
			if cerr, ok := err.(Error); ok {
				assert.Equal(test.expectedErrCode, cerr.Code)
			} else {
				assert.NoError(err)
			}
			assert.Equal(test.expected, result != nil, "expects Find() to return an event")
		})
	}
}
