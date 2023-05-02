package agent

import (
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/stretchr/testify/assert"
)

func TestGetAgentEntity(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		name              string
		agent             *Agent
		expectedAgentName string
	}{
		{
			name: "The agent has no entity",
			agent: &Agent{
				config: &Config{
					AgentName:               "foo",
					Namespace:               "default",
					DeregistrationPipelines: []string{"core/v2.Pipeline.slack"},
				},
			},
			expectedAgentName: "foo",
		},
		{
			name: "The agent has an entity",
			agent: &Agent{
				entityConfig: corev3.FixtureEntityConfig("bar"),
				config: &Config{
					AgentName:               "bar",
					Namespace:               "default",
					DeregistrationPipelines: []string{"core/v2.Pipeline.slack"},
				},
			},
			expectedAgentName: "bar",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.agent.systemInfo = &corev2.System{}

			entity := tc.agent.getAgentEntity()
			assert.Equal(tc.expectedAgentName, entity.Name)
		})
	}
}

func TestGetEntities(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		name              string
		agent             *Agent
		event             *corev2.Event
		expectedAgentName string
		expectedSource    string
	}{
		{
			name: "The provided event has no entity",
			agent: &Agent{
				entityConfig: corev3.FixtureEntityConfig("foo"),
				config: &Config{
					Namespace: "default",
					AgentName: "foo",
				},
			},
			event: &corev2.Event{
				Check: corev2.FixtureCheck("check_cpu"),
			},
			expectedAgentName: "foo",
			expectedSource:    "",
		},
		{
			name: "The provided event has an entity",
			agent: &Agent{
				config: &Config{
					Namespace: "default",
					AgentName: "agent_entity",
				},
			},
			event:             corev2.FixtureEvent("proxy_entity", "check_cpu"),
			expectedAgentName: "agent_entity",
			expectedSource:    "proxy_entity",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.agent.systemInfo = &corev2.System{}

			tc.agent.getEntities(tc.event)
			assert.Equal(tc.expectedAgentName, tc.event.Entity.Name)
			assert.Equal(tc.expectedSource, tc.event.Check.ProxyEntityName)
		})
	}
}
