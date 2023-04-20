package globalid

import (
	v2 "github.com/sensu/core/v2"
	//
	// Checks
	//
)

var checkName = "checks"

// CheckTranslator global ID resource
var CheckTranslator = commonTranslator{
	name:		checkName,
	encodeFunc:	standardEncoder(checkName, "Name"),
	decodeFunc:	standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*v2.CheckConfig)
		return ok
	},
}

// Register entity encoder/decoder
func init()	{ RegisterTranslator(CheckTranslator) }
