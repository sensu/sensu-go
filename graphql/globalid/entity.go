package globalid

import "github.com/sensu/sensu-go/types"

//
// Entity
//

var entityName = "entities"

// EntityResource global ID resource
var EntityResource = commonResource{
	name:       entityName,
	encodeFunc: standardEncoder(entityName, "ID"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.Entity)
		return ok
	},
}

// Register entity encoder/decoder
func init() { registerResource(EntityResource) }
