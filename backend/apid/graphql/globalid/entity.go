package globalid

import (
	v2 "github.com/sensu/core/v2"
)

//
// Entity
//

var entityName = "entities"

// EntityTranslator global ID resource
var EntityTranslator = commonTranslator{
	name:       entityName,
	encodeFunc: standardEncoder(entityName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*v2.Entity)
		return ok
	},
}

// Register entity encoder/decoder
func init() { RegisterTranslator(EntityTranslator) }
