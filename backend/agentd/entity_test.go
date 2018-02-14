package agentd

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetProxyEntity(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	store.On("GetEntityByID", mock.Anything, "bar").Return(types.FixtureEntity("bar"), nil)

	var nilEntity *types.Entity
	store.On("GetEntityByID", mock.Anything, "baz").Return(nilEntity, nil)
	store.On("UpdateEntity", mock.Anything, mock.Anything).Once().Return(nil)

	store.On("GetEntityByID", mock.Anything, "quux").Return(nilEntity, errors.New("error"))

	store.On("GetEntityByID", mock.Anything, "qux").Return(nilEntity, nil)
	store.On("UpdateEntity", mock.Anything, mock.Anything).Once().Return(errors.New("error"))

	testCases := []struct {
		name           string
		event          *types.Event
		expectedError  bool
		expectedEntity string
	}{
		{
			name:           "The event has no proxy entity",
			event:          types.FixtureEvent("foo", "check_cpu"),
			expectedError:  false,
			expectedEntity: "foo",
		},
		{
			name: "The event has a proxy entity with a corresponding entity",
			event: &types.Event{
				Check: &types.Check{
					ProxyEntityID: "bar",
				},
				Entity: types.FixtureEntity("foo"),
			},
			expectedError:  false,
			expectedEntity: "bar",
		},
		{
			name: "The event has a proxy entity with no corresponding entity",
			event: &types.Event{
				Check: &types.Check{
					ProxyEntityID: "baz",
				},
				Entity: types.FixtureEntity("foo"),
			},
			expectedError:  false,
			expectedEntity: "baz",
		},
		{
			name: "The proxy entity can't be queried",
			event: &types.Event{
				Check: &types.Check{
					ProxyEntityID: "quux",
				},
				Entity: types.FixtureEntity("foo"),
			},
			expectedError: true,
		},
		{
			name: "The proxy entity can't be created",
			event: &types.Event{
				Check: &types.Check{
					ProxyEntityID: "qux",
				},
				Entity: types.FixtureEntity("foo"),
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := getProxyEntity(tc.event, store)
			testutil.CompareError(err, tc.expectedError, t)

			if tc.expectedEntity != "" {
				assert.Equal(tc.expectedEntity, tc.event.Entity.ID)
			}
		})
	}
}
