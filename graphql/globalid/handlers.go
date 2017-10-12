package globalid

import "github.com/sensu/sensu-go/types"

//
// Handler
//

var handlerName = "handlers"

// HandlerResource global ID resource
var HandlerResource = commonResource{
	name:       handlerName,
	encodeFunc: standardEncoder(handlerName, "ID"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.Handler)
		return ok
	},
}

// Register handler encoder/decoder
func init() { registerResource(HandlerResource) }
