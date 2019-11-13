package globalid

import (
	"context"

	"github.com/sensu/sensu-go/types"
)

//
// Namespaces
//

var namespaceName = "namespaces"

// NamespaceTranslator global ID resource
var NamespaceTranslator = commonTranslator{
	name:       namespaceName,
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.Namespace)
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
func init() { RegisterTranslator(NamespaceTranslator) }
