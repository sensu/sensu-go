package globalid

import "github.com/sensu/sensu-go/types"

//
// Environments
//

var environmentName = "environments"

// EnvironmentTranslator global ID resource
var EnvironmentTranslator = commonTranslator{
	name:       environmentName,
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.Environment)
		return ok
	},

	//
	// Example output:
	//
	//   srn:environments:myorg:myenv
	//   srn:environments:myotherorg:myproductionenv
	//
	encodeFunc: func(record interface{}) Components {
		env, ok := record.(*types.Environment)
		if !ok {
			return nil
		}

		components := StandardComponents{
			resource:        environmentName,
			organization:    env.Organization,
			uniqueComponent: env.Name,
		}
		return &components
	},
}

// Register entity encoder/decoder
func init() { registerTranslator(EnvironmentTranslator) }
