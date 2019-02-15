package agent

import (
	"encoding/json"
	"fmt"
	"testing"

	time "github.com/echlebek/timeproxy"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types/v1"
)

func TestTranslateToEvent(t *testing.T) {
	agent := &Agent{
		config: &Config{
			AgentName:     "test-agent",
			Namespace:     "test-namespace",
			Subscriptions: []string{"default"},
			User:          "test-user",
		},
		systemInfo: &corev2.System{},
	}
	tests := []struct {
		Name      string
		Input     string
		ExpOutput *corev2.Event
		ExpError  bool
	}{
		{
			Name:  "check from docs",
			Input: `{"name": "check-mysql-status", "output": "error!", "status": 1, "handlers": ["slack"]}`,
			ExpOutput: &corev2.Event{
				Check: &corev2.Check{
					ObjectMeta: corev2.ObjectMeta{
						Name:      "check-mysql-status",
						Namespace: "test-namespace",
					},
					Output:   "error!",
					Status:   1,
					Handlers: []string{"slack"},
				},
				Entity: &corev2.Entity{
					ObjectMeta: corev2.ObjectMeta{
						Name:      "test-agent",
						Namespace: "test-namespace",
					},
					EntityClass:   corev2.EntityAgentClass,
					Subscriptions: []string{"default"},
					User:          "test-user",
					LastSeen:      time.Now().Unix(),
				},
			},
		},
		{
			Name:  "check with client",
			Input: `{"name": "check-mysql-status", "output": "error!", "status": 1, "handlers": ["slack"], "client": "foobar"}`,
			ExpOutput: &corev2.Event{
				Check: &corev2.Check{
					ObjectMeta: corev2.ObjectMeta{
						Name:      "check-mysql-status",
						Namespace: "test-namespace",
					},
					Output:   "error!",
					Status:   1,
					Handlers: []string{"slack"},
				},
				Entity: &corev2.Entity{
					ObjectMeta: corev2.ObjectMeta{
						Name:      "foobar",
						Namespace: "test-namespace",
					},
					EntityClass: corev2.EntityProxyClass,
				},
			},
		},
		{
			Name:  "check with deprecated handler attr",
			Input: `{"name": "check-mysql-status", "output": "error!", "status": 1, "handler": "slack", "client": "foobar"}`,
			ExpOutput: &corev2.Event{
				Check: &corev2.Check{
					ObjectMeta: corev2.ObjectMeta{
						Name:      "check-mysql-status",
						Namespace: "test-namespace",
					},
					Output:   "error!",
					Status:   1,
					Handlers: []string{"slack"},
				},
				Entity: &corev2.Entity{
					ObjectMeta: corev2.ObjectMeta{
						Name:      "foobar",
						Namespace: "test-namespace",
					},
					EntityClass: corev2.EntityProxyClass,
				},
			},
		},
		{
			Name:  "check with deprecated handler attr and new handlers attr",
			Input: `{"name": "check-mysql-status", "output": "error!", "status": 1, "handler": "poop", "handlers": ["slack"], "client": "foobar"}`,
			ExpOutput: &corev2.Event{
				Check: &corev2.Check{
					ObjectMeta: corev2.ObjectMeta{
						Name:      "check-mysql-status",
						Namespace: "test-namespace",
					},
					Output:   "error!",
					Status:   1,
					Handlers: []string{"slack"},
				},
				Entity: &corev2.Entity{
					ObjectMeta: corev2.ObjectMeta{
						Name:      "foobar",
						Namespace: "test-namespace",
					},
					EntityClass: corev2.EntityProxyClass,
				},
			},
		},
		{
			Name:     "missing name",
			Input:    `{"output": "error!", "status": 1, "handler": "poop", "handlers": ["slack"], "client": "foobar"}`,
			ExpError: true,
		},
		{
			Name:     "missing output",
			Input:    `{"name": "check-mysql-status", "status": 1, "handler": "poop", "handlers": ["slack"], "client": "foobar"}`,
			ExpError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var result v1.CheckResult
			if err := json.Unmarshal([]byte(test.Input), &result); err != nil {
				t.Fatal(err)
			}
			event := &corev2.Event{}
			err := translateToEvent(agent, result, event)
			if got, want := (err != nil), test.ExpError; got != want {
				if !want {
					t.Fatal(err)
				}
				t.Error("expected non-nil error")
			}

			if err != nil {
				return
			}

			if got, want := fmt.Sprintf("%v", event), fmt.Sprintf("%v", test.ExpOutput); got != want {
				t.Errorf("bad output: got %v, want %v", got, want)
			}
		})
	}
}
