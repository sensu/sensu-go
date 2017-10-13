package globalid

import (
	"strconv"

	"github.com/sensu/sensu-go/types"
)

//
// Events
//

const eventName = "events"
const eventCheckType = "check"
const eventMetricType = "metric"

// EventComponents adds methods to easily access unique elements of event.
type EventComponents struct{ StandardComponents }

// newEventComponents instantiates new EventComponents composite.
func newEventComponents(components StandardComponents) Components {
	return EventComponents{components}
}

// EntityName method returns first element of global ID's uniqueComponents.
func (n EventComponents) EntityName() string {
	return n.uniqueComponents[0]
}

// CheckName method returns second element of global ID's uniqueComponents IF
// the event is associated with a check otherwise returns an empty string.
func (n EventComponents) CheckName() string {
	if n.ResourceType() == eventCheckType {
		return n.uniqueComponents[1]
	}
	return ""
}

// MetricID method returns second element of global ID's uniqueComponents IF
// the event is associated with a metric otherwise returns an empty string.
func (n EventComponents) MetricID() string {
	if n.ResourceType() == eventMetricType {
		return n.uniqueComponents[1]
	}
	return ""
}

// Timestamp method returns final element of global ID's uniqueComponents.
func (n EventComponents) Timestamp() int64 {
	timestampStr := n.uniqueComponents[2]
	timestamp, _ := strconv.ParseInt(timestampStr, 10, 64)
	return timestamp
}

// EventResource global ID resource
var EventResource = commonResource{
	name:       eventName,
	decodeFunc: newEventComponents,
	encodeFunc: func(record interface{}) Components {
		event := record.(*types.Event)
		components := encodeEvent(event)
		return components
	},
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.Event)
		return ok
	},
}

// Register event encoder/decoder
func init() { registerResource(EventResource) }

//
// Example output:
//
//   srn:events:myorg:myenv:check/my-entity:my-check:1257894000
//   srn:events:myorg:myenv:metric/my-entity:my-metric:1257894000
//
func encodeEvent(event *types.Event) StandardComponents {
	components := StandardComponents{}
	components.resource = eventName
	addMultitenantFields(&components, event.Entity)

	timestamp := strconv.FormatInt(event.Timestamp, 10)
	if event.Check != nil {
		components.resourceType = eventCheckType
		components.uniqueComponents = []string{
			event.Entity.ID,
			event.Check.Config.Name,
			timestamp,
		}
	} else if event.Metrics != nil {
		components.resourceType = eventMetricType
		components.uniqueComponents = []string{
			event.Entity.ID,
			"1234", // event.Metrics.ID,
			timestamp,
		}
	}

	return components
}
