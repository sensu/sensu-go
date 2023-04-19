package globalid

import (
	v2 "github.com/sensu/core/v2"
	//
	// Cluster Roles
	//
)

var clusterRoleName = "clusterroles"

// ClusterRoleTranslator global ID resource
var ClusterRoleTranslator = commonTranslator{
	name:       clusterRoleName,
	encodeFunc: standardEncoder(clusterRoleName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*v2.ClusterRole)
		return ok
	},
}

// Register entity encoder/decoder
func init() { RegisterTranslator(ClusterRoleTranslator) }
