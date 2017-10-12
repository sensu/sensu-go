package globalid

import "github.com/sensu/sensu-go/types"

//
// Environments
//

var environmentName = "environments"

// EnvironmentResource global ID resource
var EnvironmentResource = commonResource{
	name:       environmentName,
	encodeFunc: standardEncoder(environmentName, "Name"), // TODO: Include org.
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.Environment)
		return ok
	},
}

// Register entity encoder/decoder
func init() { registerResource(EnvironmentResource) }
