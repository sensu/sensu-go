package globalid

import (
	"context"

	v2 "github.com/sensu/core/v2"
)

//
// Namespaces
//

var namespaceName = "namespaces"

// NamespaceTranslator global ID resource
var NamespaceTranslator = commonTranslator{
	name:		namespaceName,
	decodeFunc:	standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*v2.Namespace)
		return ok
	},

	//
	// Example output:
	//
	//   srn:namespaces:myns
	//   srn:namespaces:myns
	//
	encodeFunc: func(ctx context.Context, record interface{}) Components {
		components := Encode(ctx, record)
		components.resource = namespaceName
		return components
	},
}

// Register entity encoder/decoder
func init()	{ RegisterTranslator(NamespaceTranslator) }
