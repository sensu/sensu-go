package globalid

import "github.com/sensu/sensu-go/types"

//
// Environments
//

var environmentName = "environments"

// EnvironmentTranslator global ID resource
var EnvironmentTranslator = commonTranslator{
	name:       environmentName,
	encodeFunc: standardEncoder(environmentName, "Name"), // TODO: Include org.
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.Environment)
		return ok
	},
}

// Register entity encoder/decoder
func init() { registerTranslator(EnvironmentTranslator) }
