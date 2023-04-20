package globalid

import (
	v2 "github.com/sensu/core/v2"
	//
	// Silences
	//
)

var silenceName = "silences"

// SilenceTranslator global ID resource
var SilenceTranslator = commonTranslator{
	name:		silenceName,
	encodeFunc:	standardEncoder(silenceName, "Name"),
	decodeFunc:	standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*v2.Silenced)
		return ok
	},
}

// Register silence encoder/decoder
func init()	{ RegisterTranslator(SilenceTranslator) }
