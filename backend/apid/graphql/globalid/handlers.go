package globalid

import "github.com/sensu/sensu-go/types"

//
// Handler
//

var handlerName = "handlers"

// HandlerTranslator global ID resource
var HandlerTranslator = commonTranslator{
	name:       handlerName,
	encodeFunc: standardEncoder(handlerName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.Handler)
		return ok
	},
}

// Register handler encoder/decoder
func init() { RegisterTranslator(HandlerTranslator) }
