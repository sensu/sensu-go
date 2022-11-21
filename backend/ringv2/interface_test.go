package ringv2

import "testing"

func TestSubscriptionValidate(t *testing.T) {
	tests := []struct {
		Name         string
		Subscription Subscription
		WantErr      bool
	}{
		{
			Name: "no name",
			Subscription: Subscription{
				Items:            5,
				IntervalSchedule: 10,
			},
			WantErr: true,
		},
		{
			Name: "no items",
			Subscription: Subscription{
				Name:             "yes",
				IntervalSchedule: 10,
			},
			WantErr: true,
		},
		{
			Name: "no schedule",
			Subscription: Subscription{
				Name:  "hello",
				Items: 5,
			},
			WantErr: true,
		},
		{
			Name: "schedule conflict",
			Subscription: Subscription{
				Name:             "hello",
				Items:            5,
				IntervalSchedule: 10,
				CronSchedule:     "* * * * *",
			},
			WantErr: true,
		},
		{
			Name: "malformed cron",
			Subscription: Subscription{
				Name:         "hello",
				Items:        5,
				CronSchedule: "next tuesday",
			},
			WantErr: true,
		},
		{
			Name: "good interval",
			Subscription: Subscription{
				Name:             "hello",
				Items:            5,
				IntervalSchedule: 10,
			},
		},
		{
			Name: "good cron",
			Subscription: Subscription{
				Name:         "hello",
				Items:        5,
				CronSchedule: "* * * * *",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := test.Subscription.Validate()
			if test.WantErr {
				if err == nil {
					t.Fatal("expected non-nil error")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
