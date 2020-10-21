package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/v2/storetest"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
)

func TestNewEntityController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	storev2 := &storetest.Store{}
	actions := NewEntityController(store, storev2)

	assert.NotNil(actions)
	assert.Equal(store, actions.store)
}

func TestEntityFind(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		record          *types.Entity
		argument        string
		expected        bool
		expectedErrCode ErrCode
	}{
		{
			name:            "no argument given",
			ctx:             defaultCtx,
			argument:        "",
			expected:        false,
			expectedErrCode: NotFound,
		},
		{
			name:            "found",
			ctx:             defaultCtx,
			record:          types.FixtureEntity("entity1"),
			argument:        "entity1",
			expected:        true,
			expectedErrCode: 0,
		},
		{
			name:            "not found",
			ctx:             defaultCtx,
			record:          nil,
			argument:        "missing",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		storev2 := &storetest.Store{}
		actions := NewEntityController(store, storev2)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetEntityByName", tc.ctx, mock.Anything, mock.Anything).
				Return(tc.record, nil)

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

func TestEntityList(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	testCases := []struct {
		name        string
		ctx         context.Context
		records     []*types.Entity
		storeErr    error
		expectedLen int
		expectedErr error
	}{
		{
			name:        "no results",
			ctx:         defaultCtx,
			records:     []*types.Entity{},
			expectedLen: 0,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name: "with results",
			ctx:  defaultCtx,
			records: []*types.Entity{
				types.FixtureEntity("entity1"),
				types.FixtureEntity("entity2"),
			},
			expectedLen: 2,
			storeErr:    nil,
			expectedErr: nil,
		},
		{
			name:        "store failure",
			ctx:         defaultCtx,
			records:     nil,
			expectedLen: 0,
			storeErr:    errors.New(""),
			expectedErr: NewError(InternalErr, errors.New("")),
		},
	}

	for _, tc := range testCases {
		s := &mockstore.MockStore{}
		s2 := &storetest.Store{}
		actions := NewEntityController(s, s2)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			pred := &store.SelectionPredicate{}
			s.On("GetEntities", tc.ctx, pred).Return(tc.records, tc.storeErr)

			// Exec Query
			results, err := actions.List(tc.ctx, pred)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestEntityCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badEntity := types.FixtureEntity("badentity")
	badEntity.Name = ""

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Entity
		fetchResult     *types.Entity
		fetchErr        error
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Created",
			ctx:         defaultCtx,
			argument:    types.FixtureEntity("foo"),
			fetchResult: nil,
			fetchErr:    nil,
			createErr:   nil,
			expectedErr: false,
		},
		{
			name:            "store err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureEntity("foo"),
			fetchResult:     nil,
			fetchErr:        nil,
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "store err on fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureEntity("foo"),
			fetchResult:     types.FixtureEntity("foo"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Validation error",
			ctx:             defaultCtx,
			argument:        badEntity,
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		storev2 := &storetest.Store{}
		actions := NewEntityController(store, storev2)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetEntityByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateEntity", mock.Anything, mock.Anything).
				Return(tc.createErr)

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

func TestEntityCreateOrReplace(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	agentEntity := corev2.FixtureEntity("agent-entity")
	agentEntity.EntityClass = corev2.EntityAgentClass

	proxyEntity := corev2.FixtureEntity("proxy-entity")
	proxyEntity.EntityClass = corev2.EntityProxyClass

	badEntity := corev2.FixtureEntity("badentity")
	badEntity.Name = ""

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Entity
		exists          bool
		createErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "agent entity",
			ctx:         defaultCtx,
			argument:    agentEntity,
			exists:      true,
			createErr:   nil,
			expectedErr: false,
		},
		{
			name:            "agent entity, store failure",
			ctx:             defaultCtx,
			argument:        agentEntity,
			exists:          true,
			createErr:       NewError(InternalErr, errors.New("some error")),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:        "proxy entity",
			ctx:         defaultCtx,
			argument:    proxyEntity,
			createErr:   nil,
			expectedErr: false,
		},
		{
			name:            "proxy entity, store failure",
			ctx:             defaultCtx,
			argument:        proxyEntity,
			createErr:       NewError(InternalErr, errors.New("some error")),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Validation error",
			ctx:             defaultCtx,
			argument:        badEntity,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
		{
			name:        "entity that does not exist gets properly created",
			ctx:         defaultCtx,
			argument:    agentEntity,
			exists:      false,
			expectedErr: false,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		storev2 := &storetest.Store{}
		actions := NewEntityController(store, storev2)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("UpdateEntity", mock.Anything, mock.Anything).
				Return(tc.createErr)

			storev2.
				On("CreateOrUpdate", mock.Anything, mock.Anything).
				Return(tc.createErr)

			if tc.exists {
				store.
					On("GetEntityByName", mock.Anything, mock.Anything).
					Return(tc.argument, nil)
			} else {
				var nilEntity *corev2.Entity
				store.
					On("GetEntityByName", mock.Anything, mock.Anything).
					Return(nilEntity, nil)
			}

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
