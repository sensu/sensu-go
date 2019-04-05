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

func TestNewEventFilterController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	ctl := NewEventFilterController(store)
	assert.NotNil(ctl)
	assert.Equal(store, ctl.store)
}

func TestEventFilterCreateOrReplace(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badMut := types.FixtureEventFilter("bad")
	badMut.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	tests := []struct {
		name            string
		ctx             context.Context
		argument        *types.EventFilter
		fetchResult     *types.EventFilter
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureEventFilter("sleepy"),
			expectedErr: false,
		},
		{
			name:        "Already Exists",
			ctx:         defaultCtx,
			argument:    types.FixtureEventFilter("sleepy"),
			fetchResult: types.FixtureEventFilter("sleepy"),
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
		ctl := NewEventFilterController(store)

		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			store.On("GetEventFilterByName", mock.Anything, mock.Anything).
				Return(test.fetchResult, test.fetchErr)

			store.On("UpdateEventFilter", mock.Anything, mock.Anything).Return(test.createErr)

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

func TestEventFilterCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badMut := types.FixtureEventFilter("bad")
	badMut.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	tests := []struct {
		name            string
		ctx             context.Context
		argument        *types.EventFilter
		fetchResult     *types.EventFilter
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureEventFilter("sleepy"),
			expectedErr: false,
		},
		{
			name:            "Already Exists",
			ctx:             defaultCtx,
			argument:        types.FixtureEventFilter("sleepy"),
			fetchResult:     types.FixtureEventFilter("sleepy"),
			expectedErr:     true,
			expectedErrCode: AlreadyExistsErr,
		},
		{
			name:            "store Err on Fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureEventFilter("grumpy"),
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
		ctl := NewEventFilterController(store)

		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			store.On("GetEventFilterByName", mock.Anything, mock.Anything).
				Return(test.fetchResult, test.fetchErr)

			store.On("UpdateEventFilter", mock.Anything, mock.Anything).Return(test.createErr)

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

func TestEventFilterDestroy(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		filterName      string
		fetchResult     *types.EventFilter
		fetchErr        error
		deleteErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Deleted",
			ctx:         defaultCtx,
			filterName:  "filter1",
			fetchResult: types.FixtureEventFilter("filter1"),
			expectedErr: false,
		},
		{
			name:            "Does Not Exist",
			ctx:             defaultCtx,
			filterName:      "filter1",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "store Err on Delete",
			ctx:             defaultCtx,
			filterName:      "filter1",
			fetchResult:     types.FixtureEventFilter("filter1"),
			deleteErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "store Err on Fetch",
			ctx:             defaultCtx,
			filterName:      "filter1",
			fetchResult:     types.FixtureEventFilter("filter1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewEventFilterController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetEventFilterByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("DeleteEventFilterByName", mock.Anything, "filter1").
				Return(tc.deleteErr)

			// Exec Query
			err := actions.Destroy(tc.ctx, tc.filterName)

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

func TestEventFilterList(t *testing.T) {
	readCtx := context.Background()

	testCases := []struct {
		name        string
		ctx         context.Context
		records     []*types.EventFilter
		storeErr    error
		expectedLen int
		expectedErr error
	}{
		{
			name:        "No EventFilters",
			ctx:         readCtx,
			records:     nil,
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "EventFilters",
			ctx:  readCtx,
			records: []*types.EventFilter{
				types.FixtureEventFilter("homer"),
				types.FixtureEventFilter("bart"),
			},
			expectedLen: 2,
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
		actions := NewEventFilterController(s)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			pred := &store.SelectionPredicate{}
			s.On("GetEventFilters", tc.ctx, pred).Return(tc.records, tc.storeErr)

			// Exec Query
			results, err := actions.List(tc.ctx, pred)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestEventFilterFind(t *testing.T) {
	readCtx := context.Background()

	tests := []struct {
		name            string
		ctx             context.Context
		filter          *types.EventFilter
		filterName      string
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "No filter name",
			ctx:             readCtx,
			expected:        false,
			expectedErrCode: InvalidArgument,
		},
		{
			name:            "Found",
			ctx:             readCtx,
			filter:          types.FixtureEventFilter("abe"),
			filterName:      "abe",
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "Not Found",
			ctx:             readCtx,
			filter:          nil,
			filterName:      "fox mulder",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := &mockstore.MockStore{}
			ctl := NewEventFilterController(store)

			// Mock store methods
			store.On("GetEventFilterByName", test.ctx, test.filterName).
				Return(test.filter, nil)

			assert := assert.New(t)
			result, err := ctl.Find(test.ctx, test.filterName)
			if cerr, ok := err.(Error); ok {
				assert.Equal(test.expectedErrCode, cerr.Code)
			} else {
				assert.NoError(err)
			}
			assert.Equal(test.expected, result != nil, "expects Find() to return an event")
		})
	}
}
