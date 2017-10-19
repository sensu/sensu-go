package eventd

import (
	"context"
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// test cases:
//		not silenced with subscriptions
func TestIsSilenced(t *testing.T) {
	testCases := []struct {
		description             string
		event                   *types.Event
		silencedSubscriptions   []*types.Silenced
		expectedSilencedEntries []string
	}{
		{
			"silenced with one subscription",
			&types.Event{
				Entity: &types.Entity{
					Subscriptions: []string{"subscription1", "subscription2"},
				},
				Check: &types.Check{
					Config: &types.CheckConfig{},
					Status: 1,
				},
			},
			[]*types.Silenced{
				&types.Silenced{
					ID:           "subscription1:check1",
					Subscription: "subscription1",
					CheckName:    "check1",
				},
			},
			[]string{"subscription1"},
		},
		{
			"silenced with multiple subscriptions",
			&types.Event{
				Entity: &types.Entity{
					Subscriptions: []string{"subscription1", "subscription2", "subscription3"},
				},
				Check: &types.Check{
					Config: &types.CheckConfig{},
					Status: 1,
				},
			},
			[]*types.Silenced{
				&types.Silenced{
					ID:           "subscription1:check1",
					Subscription: "subscription1",
					CheckName:    "check1",
				},
				&types.Silenced{
					ID:           "subscription2:check2",
					Subscription: "subscription2",
					CheckName:    "check2",
				},
			},
			[]string{"subscription1", "subscription2"},
		},
		{
			"silenced with no silenced events",
			&types.Event{
				Entity: &types.Entity{
					Subscriptions: []string{"subscription1", "subscription2"},
				},
				Check: &types.Check{
					Config: &types.CheckConfig{},
					Status: 1,
				},
			},
			[]*types.Silenced{},
			nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), types.OrganizationKey, "default")
			ctx = context.WithValue(ctx, types.EnvironmentKey, "default")

			mockStore := &mockstore.MockStore{}
			mockStore.On(
				"GetSilencedEntriesBySubscription",
				mock.Anything,
			).Return(tc.silencedSubscriptions, nil)

			fmt.Println(tc.event)
			result := getSilenced(ctx, tc.event, mockStore)
			assert.Nil(t, result)
			assert.Equal(t, tc.expectedSilencedEntries, tc.event.Silenced)
		})
	}
}
