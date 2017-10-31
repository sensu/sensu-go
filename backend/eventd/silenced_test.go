package eventd

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// test cases:
// 	silenced check
func TestIsSilenced(t *testing.T) {
	testCases := []struct {
		description             string
		event                   *types.Event
		silencedSubscriptions   []*types.Silenced
		silencedChecks          []*types.Silenced
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
					Check:        "check1",
				},
			},
			[]*types.Silenced{},
			[]string{"subscription1:check1"},
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
					Check:        "check1",
				},
				&types.Silenced{
					ID:           "subscription2:check2",
					Subscription: "subscription2",
					Check:        "check2",
				},
			},
			[]*types.Silenced{},
			[]string{"subscription1:check1", "subscription2:check2"},
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
			[]*types.Silenced{},
			nil,
		},
		{
			"silenced with check and subscription",
			&types.Event{
				Entity: &types.Entity{
					Subscriptions: []string{"subscription1"},
				},
				Check: &types.Check{
					Config: &types.CheckConfig{
						Name: "check2",
					},
					Status: 1,
				},
			},
			[]*types.Silenced{
				&types.Silenced{
					ID:           "subscription1:check1",
					Subscription: "subscription1",
					Check:        "check1",
				},
			},
			[]*types.Silenced{
				&types.Silenced{
					ID:           "subscription2:check2",
					Subscription: "subscription2",
					Check:        "check2",
				},
			},
			[]string{"subscription1:check1", "subscription2:check2"},
		},
		{
			"silenced with duplicate check and subscription",
			&types.Event{
				Entity: &types.Entity{
					Subscriptions: []string{"subscription1", "subscription2"},
				},
				Check: &types.Check{
					Config: &types.CheckConfig{
						Name: "check1",
					},
					Status: 1,
				},
			},
			[]*types.Silenced{
				&types.Silenced{
					ID:           "subscription1:check1",
					Subscription: "subscription1",
					Check:        "check1",
				},
				&types.Silenced{
					ID:           "subscription2:check2",
					Subscription: "subscription2",
					Check:        "check2",
				},
			},
			[]*types.Silenced{
				&types.Silenced{
					ID:           "subscription2:check2",
					Subscription: "subscription2",
					Check:        "check2",
				},
			},
			[]string{"subscription1:check1", "subscription2:check2"},
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

			mockStore.On(
				"GetSilencedEntriesByCheckName",
				mock.Anything,
			).Return(tc.silencedChecks, nil)

			result := getSilenced(ctx, tc.event, mockStore)
			assert.Nil(t, result)
			assert.Equal(t, tc.expectedSilencedEntries, tc.event.Silenced)
		})
	}
}
