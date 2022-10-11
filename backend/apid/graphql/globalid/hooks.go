package globalid

import corev2 "github.com/sensu/core/v2"

//
// Hooks
//

var hookName = "hooks"

// HookTranslator global ID resource
var HookTranslator = commonTranslator{
	name:       hookName,
	encodeFunc: standardEncoder(hookName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*corev2.HookConfig)
		return ok
	},
}

// Register entity encoder/decoder
func init() { RegisterTranslator(HookTranslator) }
