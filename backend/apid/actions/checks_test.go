package actions

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/testing/mockqueue"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewCheckController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.V2MockStore{}
	actions := NewCheckController(store, queue.NewMemoryClient())

	assert.NotNil(actions)
	assert.Equal(store, actions.store)
}

func TestCheckAdhoc(t *testing.T) {
	t.Skip("skip")
	defaultCtx := testutil.NewContext(
		testutil.ContextWithNamespace("default"),
	)

	badCheck := corev2.FixtureCheckConfig("check1")
	badCheck.Name = "!@#!#$@#^$%&$%&$&$%&%^*%&(%@###"

	testCases := []struct {
		name		string
		ctx		context.Context
		argument	*corev2.AdhocRequest
		fetchResult	*corev2.CheckConfig
		checkName	string
		fetchErr	error
		queueErr	error
		expectedErr	bool
		expectedErrCode	ErrCode
	}{
		{
			name:		"Queued",
			ctx:		defaultCtx,
			argument:	corev2.FixtureAdhocRequest("check1", []string{"subscription1", "subscription2"}),
			fetchResult:	corev2.FixtureCheckConfig("check1"),
			checkName:	"check1",
			fetchErr:	nil,
			queueErr:	nil,
			expectedErr:	false,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.V2MockStore{}
		cs := new(mockstore.ConfigStore)
		store.On("GetConfigStore").Return(cs)
		queue := &mockqueue.MockQueue{}
		actions := NewCheckController(store, queue)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mock store methods
			cs.
				On("Get", mock.Anything, mock.Anything).
				Return(mockstore.Wrapper[*corev2.CheckConfig]{Value: tc.fetchResult}, tc.fetchErr)
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
