package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewEntityController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewEntityController(store)

	assert.NotNil(actions)
	assert.Equal(store, actions.Store)
	assert.NotNil(actions.Policy)
}

func TestEntityDestroy(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEntity, types.RulePermDelete),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEntity, types.RulePermCreate),
		),
	)

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        string
		fetchResult     *types.Entity
		fetchErr        error
		deleteErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "deleted",
			ctx:         defaultCtx,
			argument:    "entity1",
			fetchResult: types.FixtureEntity("entity1"),
			expectedErr: false,
		},
		{
			name:            "does not exist",
			ctx:             defaultCtx,
			argument:        "entity1",
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "store err on delete",
			ctx:             defaultCtx,
			argument:        "entity1",
			fetchResult:     types.FixtureEntity("entity1"),
			deleteErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "store err on fetch",
			ctx:             defaultCtx,
			argument:        "entity1",
			fetchResult:     types.FixtureEntity("entity1"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "no permission",
			ctx:             wrongPermsCtx,
			argument:        "entity1",
			fetchResult:     types.FixtureEntity("entity1"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewEntityController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetEntityByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("DeleteEntityByName", mock.Anything, tc.argument).
				Return(tc.deleteErr)

			// Exec Query
			err := actions.Destroy(tc.ctx, tc.argument)

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

func TestEntityFind(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEntity, types.RulePermRead),
		),
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
		{
			name: "no read permission",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeEntity, types.RulePermCreate),
			)),
			record:          types.FixtureEntity("entity1"),
			argument:        "entity1",
			expected:        false,
			expectedErrCode: NotFound,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewEntityController(store)

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

func TestEntityQuery(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEntity, types.RulePermRead),
		),
	)

	testCases := []struct {
		name        string
		ctx         context.Context
		records     []*types.Entity
		expectedLen int
		storeErr    error
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
			name: "with only create access",
			ctx: testutil.NewContext(testutil.ContextWithRules(
				types.FixtureRuleWithPerms(types.RuleTypeEntity, types.RulePermCreate),
			)),
			records: []*types.Entity{
				types.FixtureEntity("entity1"),
				types.FixtureEntity("entity2"),
			},
			expectedLen: 0,
			storeErr:    nil,
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
		store := &mockstore.MockStore{}
		actions := NewEntityController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.On("GetEntities", tc.ctx).Return(tc.records, tc.storeErr)

			// Exec Query
			results, err := actions.Query(tc.ctx)

			// Assert
			assert.EqualValues(tc.expectedErr, err)
			assert.Len(results, tc.expectedLen)
		})
	}
}

func TestEntityUpdate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEntity, types.RulePermUpdate),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEntity, types.RulePermRead),
		),
	)

	badEntity := types.FixtureEntity("badentity")
	badEntity.Name = ""

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *types.Entity
		fetchResult     *types.Entity
		fetchErr        error
		updateErr       error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Updated",
			ctx:         defaultCtx,
			argument:    types.FixtureEntity("foo"),
			fetchResult: types.FixtureEntity("foo"),
			expectedErr: false,
		},
		{
			name:            "Does not exist",
			ctx:             defaultCtx,
			argument:        types.FixtureEntity("foo"),
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store err on update",
			ctx:             defaultCtx,
			argument:        types.FixtureEntity("foo"),
			fetchResult:     types.FixtureEntity("foo"),
			updateErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store err on fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureEntity("foo"),
			fetchResult:     types.FixtureEntity("foo"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureEntity("foo"),
			fetchResult:     types.FixtureEntity("foo"),
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:            "Validation error",
			ctx:             defaultCtx,
			argument:        badEntity,
			fetchResult:     badEntity,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewEntityController(store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetEntityByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			store.
				On("UpdateEntity", mock.Anything, mock.Anything).
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

func TestEntityCreate(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEntity, types.RulePermCreate),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEntity, types.RulePermRead),
		),
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
			name:            "Store err on Create",
			ctx:             defaultCtx,
			argument:        types.FixtureEntity("foo"),
			fetchResult:     nil,
			fetchErr:        nil,
			createErr:       errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "Store err on fetch",
			ctx:             defaultCtx,
			argument:        types.FixtureEntity("foo"),
			fetchResult:     types.FixtureEntity("foo"),
			fetchErr:        errors.New("dunno"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
		{
			name:            "No permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureEntity("foo"),
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
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
		actions := NewEntityController(store)

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
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(
				types.RuleTypeEntity,
				types.RulePermCreate,
				types.RulePermUpdate,
			),
		),
	)
	wrongPermsCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeEntity, types.RulePermRead),
		),
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
			name:        "Already exists",
			ctx:         defaultCtx,
			argument:    types.FixtureEntity("foo"),
			fetchResult: types.FixtureEntity("foo"),
		},
		{
			name:            "No permission",
			ctx:             wrongPermsCtx,
			argument:        types.FixtureEntity("foo"),
			fetchResult:     nil,
			expectedErr:     true,
			expectedErrCode: PermissionDenied,
		},
		{
			name:            "Validation error",
			ctx:             defaultCtx,
			argument:        badEntity,
			expectedErr:     true,
			expectedErrCode: InvalidArgument,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewEntityController(store)

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
