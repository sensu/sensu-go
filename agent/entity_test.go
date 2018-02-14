package agent

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestGetAgentEntity(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		name               string
		agent              *Agent
		expectedAgentID    string
		extendedAttributes []byte
	}{
		{
			name: "The agent has no entity",
			agent: &Agent{
				config: &Config{
					AgentID:               "foo",
					DeregistrationHandler: "slack",
				},
			},
			expectedAgentID: "foo",
		},
		{
			name: "The agent has an entity",
			agent: &Agent{
				entity: types.FixtureEntity("bar"),
			},
			expectedAgentID: "bar",
		},
		{
			name: "The agent has extended attributes",
			agent: &Agent{
				config: &Config{
					AgentID:            "baz",
					ExtendedAttributes: []byte(`{"foo":"bar"}`),
				},
			},
			expectedAgentID:    "baz",
			extendedAttributes: []byte(`{"foo":"bar"}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entity := tc.agent.getAgentEntity()
			assert.Equal(tc.expectedAgentID, entity.ID)
			assert.Equal(tc.extendedAttributes, entity.ExtendedAttributes)
		})
	}
}

func TestGetEntities(t *testing.T) {
	assert := assert.New(t)

	testCases := []struct {
		name            string
		agent           *Agent
		event           *types.Event
		expectedAgentID string
		expectedSource  string
	}{
		{
			name: "The provided event has no entity",
			agent: &Agent{
				entity: types.FixtureEntity("foo"),
			},
			event: &types.Event{
				Check: types.FixtureCheck("check_cpu"),
			},
			expectedAgentID: "foo",
			expectedSource:  "",
		},
		{
			name: "The provided event has an entity",
			agent: &Agent{
				config: &Config{
					AgentID: "agent_entity",
				},
			},
			event:           types.FixtureEvent("proxy_entity", "check_cpu"),
			expectedAgentID: "agent_entity",
			expectedSource:  "proxy_entity",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.agent.getEntities(tc.event)
			assert.Equal(tc.expectedAgentID, tc.event.Entity.ID)
			assert.Equal(tc.expectedSource, tc.event.Check.ProxyEntityID)
		})
	}
}
