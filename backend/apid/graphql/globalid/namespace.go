package globalid

import "github.com/sensu/sensu-go/types"

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
	encodeFunc: func(record interface{}) Components {
		nsp, ok := record.(*types.Namespace)
		if !ok {
			return nil
		}

		components := StandardComponents{
			resource:        namespaceName,
			uniqueComponent: nsp.Name,
		}
		return &components
	},
}

// Register entity encoder/decoder
func init() { registerTranslator(NamespaceTranslator) }
