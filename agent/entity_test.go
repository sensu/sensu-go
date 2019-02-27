package agent

import (
	"testing"

	"github.com/sensu/sensu-go/types"
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
					AgentName:             "foo",
					DeregistrationHandler: "slack",
				},
			},
			expectedAgentName: "foo",
		},
		{
			name: "The agent has an entity",
			agent: &Agent{
				entity: types.FixtureEntity("bar"),
			},
			expectedAgentName: "bar",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.agent.systemInfo = &types.System{}

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
		event             *types.Event
		expectedAgentName string
		expectedSource    string
	}{
		{
			name: "The provided event has no entity",
			agent: &Agent{
				entity: types.FixtureEntity("foo"),
			},
			event: &types.Event{
				Check: types.FixtureCheck("check_cpu"),
			},
			expectedAgentName: "foo",
			expectedSource:    "",
		},
		{
			name: "The provided event has an entity",
			agent: &Agent{
				config: &Config{
					AgentName: "agent_entity",
				},
			},
			event:             types.FixtureEvent("proxy_entity", "check_cpu"),
			expectedAgentName: "agent_entity",
			expectedSource:    "proxy_entity",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.agent.systemInfo = &types.System{}

			tc.agent.getEntities(tc.event)
			assert.Equal(tc.expectedAgentName, tc.event.Entity.Name)
			assert.Equal(tc.expectedSource, tc.event.Check.ProxyEntityName)
		})
	}
}
