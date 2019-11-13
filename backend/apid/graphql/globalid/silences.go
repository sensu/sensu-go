package globalid

import "github.com/sensu/sensu-go/types"

//
// Silences
//

var silenceName = "silences"

// SilenceTranslator global ID resource
var SilenceTranslator = commonTranslator{
	name:       silenceName,
	encodeFunc: standardEncoder(silenceName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.Silenced)
		return ok
	},
}

// Register silence encoder/decoder
func init() { RegisterTranslator(SilenceTranslator) }
