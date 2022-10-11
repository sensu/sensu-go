package globalid

import corev2 "github.com/sensu/core/v2"

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
		_, ok := record.(*corev2.Mutator)
		return ok
	},
}

// Register entity encoder/decoder
func init() { RegisterTranslator(MutatorTranslator) }
