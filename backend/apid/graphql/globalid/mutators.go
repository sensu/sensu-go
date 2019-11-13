package globalid

import "github.com/sensu/sensu-go/types"

//
// Mutators
//

var mutatorName = "mutators"

// MutatorTranslator global ID resource
var MutatorTranslator = commonTranslator{
	name:       mutatorName,
	encodeFunc: standardEncoder(mutatorName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.Mutator)
		return ok
	},
}

// Register entity encoder/decoder
func init() { RegisterTranslator(MutatorTranslator) }
