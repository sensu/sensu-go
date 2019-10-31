package globalid

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestEncodeEvent(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()
	event := types.FixtureEvent("one", "two")
	components := encodeEvent(ctx, event)
	assert.Equal("events", components.Resource())
	assert.Equal("default", components.Namespace())
	assert.Equal("check", components.ResourceType())
	assert.NotEmpty(components.UniqueComponent())

	event.Check = nil
	event.Metrics = &types.Metrics{}
	components = encodeEvent(ctx, event)
	assert.Equal("metric", components.ResourceType())
	assert.NotEmpty(components.UniqueComponent())
}

func TestEventComponents(t *testing.T) {
	assert := assert.New(t)
	components := NewEventComponents(&StandardComponents{
		resource:        "events",
		resourceType:    "check",
		uniqueComponent: "WyJvbmUiLCJ0d28iLCIxMjM0Il0K",
	})

	assert.Equal("one", components.EntityName())
	assert.Equal("two", components.CheckName())
	assert.Empty(components.MetricID())

	components.resourceType = "metric"
	assert.Empty(components.CheckName())
	assert.Equal("two", components.MetricID())
}
