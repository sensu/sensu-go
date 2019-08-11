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
		name               string
		agent              *Agent
		event              *types.Event
		expectedEntityName string
		expectedSource     string
	}{
		{
			name: "The provided event has no entity",
			agent: &Agent{
				entity: types.FixtureEntity("foo"),
			},
			event: &types.Event{
				Check: types.FixtureCheck("check_cpu"),
			},
			expectedEntityName: "foo",
			expectedSource:     "",
		},
		{
			name: "The provided proxy event has no entity",
			agent: &Agent{
				entity: types.FixtureEntity("foo"),
			},
			event: &types.Event{
				Check: types.FixtureProxyCheck("check_cpu", "bar"),
			},
			expectedEntityName: "",
			expectedSource:     "bar",
		},
		{
			name: "The provided event has a proxy entity",
			agent: &Agent{
				config: &Config{
					AgentName: "agent_entity",
				},
			},
			event:              types.FixtureProxyEvent("proxy_entity", "check_cpu"),
			expectedEntityName: "",
			expectedSource:     "proxy_entity",
		},
		{
			name: "The provided event has a non-proxy entity",
			agent: &Agent{
				config: &Config{
					AgentName: "agent_entity",
				},
			},
			event:              types.FixtureEvent("regular_entity", "check_cpu"),
			expectedEntityName: "regular_entity",
			expectedSource:     "regular_entity",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.agent.systemInfo = &types.System{}

			tc.agent.getEntities(tc.event)
			assert.Equal(tc.expectedEntityName, tc.event.Entity.Name)
			assert.Equal(tc.expectedSource, tc.event.Check.ProxyEntityName)
		})
	}
}
