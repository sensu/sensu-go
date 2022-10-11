package globalid

import corev2 "github.com/sensu/core/v2"

//
// Asset
//

var assetName = "assets"

// AssetTranslator global ID resource
var AssetTranslator = commonTranslator{
	name:       assetName,
	encodeFunc: standardEncoder(assetName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*corev2.Asset)
		return ok
	},
}

// Register asset encoder/decoder
func init() { RegisterTranslator(AssetTranslator) }
