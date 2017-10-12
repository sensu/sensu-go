package globalid

import "github.com/sensu/sensu-go/types"

//
// Roles
//

var roleName = "roles"

// RoleResource global ID resource
var RoleResource = commonResource{
	name:       roleName,
	encodeFunc: standardEncoder(roleName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.Role)
		return ok
	},
}

// Register entity encoder/decoder
func init() { registerResource(RoleResource) }
