package globalid

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestEncodeEvent(t *testing.T) {
	assert := assert.New(t)

	event := types.FixtureEvent("one", "two")
	components := encodeEvent(event)
	assert.Equal("events", components.Resource())
	assert.Equal("default", components.Organization())
	assert.Equal("default", components.Environment())
	assert.Equal("check", components.ResourceType())
	assert.NotEmpty(components.UniqueComponent())

	event.Check = nil
	event.Metrics = &types.Metrics{}
	components = encodeEvent(event)
	assert.Equal("metric", components.ResourceType())
	assert.NotEmpty(components.UniqueComponent())
}

func TestEventComponents(t *testing.T) {
	assert := assert.New(t)
	components := newEventComponents(StandardComponents{
		resource:        "events",
		resourceType:    "check",
		uniqueComponent: "WyJvbmUiLCJ0d28iLCIxMjM0Il0K",
	}).(EventComponents)

	assert.Equal("one", components.EntityName())
	assert.Equal("two", components.CheckName())
	assert.Empty(components.MetricID())
	assert.Equal(int64(1234), components.Timestamp())

	components.resourceType = "metric"
	assert.Empty(components.CheckName())
	assert.Equal("two", components.MetricID())
}
