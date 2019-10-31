package globalid

import "github.com/sensu/sensu-go/types"

//
// Cluster Role Bindings
//
var clusterRoleBindingName = "clusterrolebindings"

// ClusterRoleBindingTranslator global ID resource
var ClusterRoleBindingTranslator = commonTranslator{
	name:       clusterRoleBindingName,
	encodeFunc: standardEncoder(clusterRoleBindingName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.ClusterRoleBinding)
		return ok
	},
}

// Register entity encoder/decoder
func init() { RegisterTranslator(ClusterRoleBindingTranslator) }
