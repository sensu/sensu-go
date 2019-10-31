package globalid

import "github.com/sensu/sensu-go/types"

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
		_, ok := record.(*types.ClusterRole)
		return ok
	},
}

// Register entity encoder/decoder
func init() { RegisterTranslator(ClusterRoleTranslator) }
