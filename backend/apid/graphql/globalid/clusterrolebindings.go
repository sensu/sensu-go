package globalid

import (
	v2 "github.com/sensu/core/v2"
	//
	// Cluster Role Bindings
	//
)

var clusterRoleBindingName = "clusterrolebindings"

// ClusterRoleBindingTranslator global ID resource
var ClusterRoleBindingTranslator = commonTranslator{
	name:		clusterRoleBindingName,
	encodeFunc:	standardEncoder(clusterRoleBindingName, "Name"),
	decodeFunc:	standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*v2.ClusterRoleBinding)
		return ok
	},
}

// Register entity encoder/decoder
func init()	{ RegisterTranslator(ClusterRoleBindingTranslator) }
