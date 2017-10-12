package globalid

import "github.com/sensu/sensu-go/types"

//
// Organizations
//

var organizationName = "organizations"

// OrganizationResource global ID resource
var OrganizationResource = commonResource{
	name:       organizationName,
	encodeFunc: standardEncoder(organizationName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*types.Organization)
		return ok
	},
}

// Register entity encoder/decoder
func init() { registerResource(OrganizationResource) }
