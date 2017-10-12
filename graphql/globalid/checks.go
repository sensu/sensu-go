package globalid

import "github.com/sensu/sensu-go/types"

//
// Checks
//

var checkName = "checks"

// CheckResource global ID resource
var CheckResource = commonResource{
	name:       checkName,
	encodeFunc: standardEncoder(checkName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.CheckConfig)
		return ok
	},
}

// Register entity encoder/decoder
func init() { registerResource(CheckResource) }
