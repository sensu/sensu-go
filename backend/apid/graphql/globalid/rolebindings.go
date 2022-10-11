package globalid

import corev2 "github.com/sensu/core/v2"

//
// Role Bindings
//
var roleBindingName = "rolebindings"

// RoleBindingTranslator global ID resource
var RoleBindingTranslator = commonTranslator{
	name:       roleBindingName,
	encodeFunc: standardEncoder(roleBindingName, "Name"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*corev2.RoleBinding)
		return ok
	},
}

// Register entity encoder/decoder
func init() { RegisterTranslator(RoleBindingTranslator) }
