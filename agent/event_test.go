package agent

import (
	"testing"

	"github.com/google/uuid"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
)

func Test_prepareEvent(t *testing.T) {
	type args struct {
		agent *Agent
		event *corev2.Event
	}
	tests := []struct {
		name          string
		args          args
		wantNamespace string
		wantErr       bool
	}{
		{
			name: "no event",
			args: args{
				agent: &Agent{},
				event: nil,
			},
			wantErr: true,
		},
		{
			name: "missing check and metrics",
			args: args{
				agent: &Agent{},
				event: &corev2.Event{
					ObjectMeta: corev2.ObjectMeta{Namespace: "default"},
				},
			},
			wantErr: true,
		},
		{
			name: "check event",
			args: args{
				agent: &Agent{
					config: &Config{
						AgentName: "agent1",
						Namespace: "default",
					},
				},
				event: &corev2.Event{
					ObjectMeta: corev2.ObjectMeta{Namespace: "default"},
					Check:      corev2.FixtureCheck("check1"),
				},
			},
			wantErr:       false,
			wantNamespace: "default",
		},
		{
			name: "metrics event",
			args: args{
				agent: &Agent{
					config: &Config{
						AgentName: "agent1",
						Namespace: "default",
					},
				},
				event: &corev2.Event{
					ObjectMeta: corev2.ObjectMeta{Namespace: "default"},
					Metrics:    corev2.FixtureMetrics(),
				},
			},
			wantErr:       false,
			wantNamespace: "default",
		},
		{
			name: "invalid check",
			args: args{
				agent: &Agent{
					config: &Config{
						Namespace: "jamespace",
					},
				},
				event: &corev2.Event{
					ObjectMeta: corev2.ObjectMeta{
						Namespace: "default",
					},
					Check: &corev2.Check{
						ProxyEntityName: "john",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "inserts missing namespace",
			args: args{
				agent: &Agent{
					config: &Config{
						Namespace: "jamespace",
					},
					entityConfig: corev3.FixtureEntityConfig("agent1"),
				},
				event: &corev2.Event{
					Check:  corev2.FixtureCheck("check1"),
					Entity: corev2.FixtureEntity("entity1"),
				},
			},
			wantErr:       false,
			wantNamespace: "jamespace",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := prepareEvent(tt.args.agent, tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("prepareEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if ns := tt.args.event.GetNamespace(); ns != tt.wantNamespace {
				t.Errorf("prepareEvent() ObjectMeta.GetNamespace() = %v, want %v", ns, tt.wantNamespace)
			}
			if id := tt.args.event.GetUUID(); id == uuid.Nil {
				t.Errorf("bad uuid: %s", id.String())
			}
			if tt.args.agent != nil && tt.args.event.Check != nil {
				if got, want := tt.args.event.Check.ProcessedBy, tt.args.agent.config.AgentName; got != want {
					t.Errorf("bad processed_by: got %q, want %q", got, want)
				}
			}
		})
	}
}
