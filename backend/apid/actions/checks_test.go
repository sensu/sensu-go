package actions

import (
	"context"
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/testing/mockqueue"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewCheckController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewCheckController(store, queue.NewMemoryGetter())

	assert.NotNil(actions)
	assert.Equal(store, actions.store)
}

func TestCheckAdhoc(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badCheck := v2.FixtureCheckConfig("check1")
	badCheck.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	testCases := []struct {
		name            string
		ctx             context.Context
		argument        *v2.AdhocRequest
		fetchResult     *v2.CheckConfig
		checkName       string
		fetchErr        error
		queueErr        error
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:        "Queued",
			ctx:         defaultCtx,
			argument:    v2.FixtureAdhocRequest("check1", []string{"subscription1", "subscription2"}),
			fetchResult: v2.FixtureCheckConfig("check1"),
			checkName:   "check1",
			fetchErr:    nil,
			queueErr:    nil,
			expectedErr: false,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		queue := &mockqueue.MockQueue{}
		getter := &mockqueue.Getter{}
		getter.On("GetQueue", mock.Anything).Return(queue)
		actions := NewCheckController(store, getter)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			store.
				On("GetCheckConfigByName", mock.Anything, mock.Anything).
				Return(tc.fetchResult, tc.fetchErr)
			queue.
				On("Enqueue", mock.Anything, mock.Anything).
				Return(tc.queueErr)

			// Exec Query
			err := actions.QueueAdhocRequest(tc.ctx, tc.checkName, tc.argument)

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
