package globalid

import corev2 "github.com/sensu/core/v2"

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
		_, ok := record.(*corev2.Handler)
		return ok
	},
}

// Register handler encoder/decoder
func init() { RegisterTranslator(HandlerTranslator) }
