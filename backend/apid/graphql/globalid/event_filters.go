package globalid

import corev2 "github.com/sensu/core/v2"

//
// Event Filters
//

var eventFilterName = "filters"

// EventFilterTranslator global ID resource
var EventFilterTranslator = commonTranslator{
	name:       eventFilterName,
	encodeFunc: standardEncoder(eventFilterName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*corev2.EventFilter)
		return ok
	},
}

// Register event filter encoder/decoder
func init() { RegisterTranslator(EventFilterTranslator) }
