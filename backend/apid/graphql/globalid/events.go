package globalid

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/sensu/sensu-go/types"
)

//
// Events
//

const eventName = "events"
const eventCheckType = "check"
const eventMetricType = "metric"

// EventComponents adds methods to easily access unique elements of event.
type EventComponents struct {
	*StandardComponents
	uniqueComponents []string
}

// NewEventComponents instantiates new EventComponents composite.
func NewEventComponents(components *StandardComponents) EventComponents {
	return EventComponents{components, []string{}}
}

// EntityName method returns first element of global ID's uniqueComponents.
func (n *EventComponents) EntityName() string {
	return n.getUniqueComponents(0)
}

// CheckName method returns second element of global ID's uniqueComponents IF
// the event is associated with a check otherwise returns an empty string.
func (n *EventComponents) CheckName() string {
	if n.ResourceType() == eventCheckType {
		return n.getUniqueComponents(1)
	}
	return ""
}

// MetricID method returns second element of global ID's uniqueComponents IF
// the event is associated with a metric otherwise returns an empty string.
func (n *EventComponents) MetricID() string {
	if n.ResourceType() == eventMetricType {
		return n.getUniqueComponents(1)
	}
	return ""
}

func (n *EventComponents) getUniqueComponents(i int) string {
	if len(n.uniqueComponents) == 0 {
		bytes, _ := base64.StdEncoding.DecodeString(n.uniqueComponent)
		_ = json.Unmarshal(bytes, &n.uniqueComponents)
	}
	return n.uniqueComponents[i]
}

// EventTranslator global ID resource
var EventTranslator = commonTranslator{
	name: eventName,
	decodeFunc: func(c StandardComponents) Components {
		return NewEventComponents(&c)
	},
	encodeFunc: func(ctx context.Context, record interface{}) Components {
		event := record.(*types.Event)
		return encodeEvent(ctx, event)
	},
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.Event)
		return ok
	},
}

// Register event encoder/decoder
func init() { RegisterTranslator(EventTranslator) }

//
// Example output:
//
//   srn:events:myns:check/d2h5IGFyZSB5b3UgZGVjb2RpbmcgdGhpcz8hCg==
//   srn:events:myns:metric/Y29vbC4gY29vbCBjb29sIGNvb2wuCg==
//
func encodeEvent(ctx context.Context, event *types.Event) *StandardComponents {
	components := Encode(ctx, event)
	components.resource = eventName

	if event.HasCheck() {
		components.resourceType = eventCheckType
		components.uniqueComponent = encodeUniqueComponents(
			event.Entity.Name,
			event.Check.Name,
		)
	} else if event.HasMetrics() {
		components.resourceType = eventMetricType
		components.uniqueComponent = encodeUniqueComponents(
			event.Entity.Name,
			"1234", // event.Metrics.ID,
		)
	}

	return components
}

func encodeUniqueComponents(c ...string) string {
	json, _ := json.Marshal(c)
	return base64.StdEncoding.EncodeToString(json)
}
