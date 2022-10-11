package globalid

import corev2 "github.com/sensu/core/v2"

//
// Cluster Roles
//
var clusterRoleName = "clusterroles"

// ClusterRoleTranslator global ID resource
var ClusterRoleTranslator = commonTranslator{
	name:       clusterRoleName,
	encodeFunc: standardEncoder(clusterRoleName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*corev2.ClusterRole)
		return ok
	},
}

// Register entity encoder/decoder
func init() { RegisterTranslator(ClusterRoleTranslator) }
