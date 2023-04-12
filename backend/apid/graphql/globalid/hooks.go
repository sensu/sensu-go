package globalid

import (
	v2 "github.com/sensu/core/v2"
	//
	// Hooks
	//
)

var hookName = "hooks"

// HookTranslator global ID resource
var HookTranslator = commonTranslator{
	name:		hookName,
	encodeFunc:	standardEncoder(hookName, "Name"),
	decodeFunc:	standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*v2.HookConfig)
		return ok
	},
}

// Register entity encoder/decoder
func init()	{ RegisterTranslator(HookTranslator) }
