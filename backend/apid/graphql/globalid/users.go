package globalid

import corev2 "github.com/sensu/core/v2"

//
// Users
//

var userName = "users"

// UserTranslator global ID resource
var UserTranslator = commonTranslator{
	name:       userName,
	encodeFunc: standardEncoder(userName, "Username"),
	decodeFunc: standardDecoder,
	isResponsibleFunc: func(record interface{}) bool {
		_, ok := record.(*corev2.User)
		return ok
	},
}

// Register user encoder/decoder
func init() { RegisterTranslator(UserTranslator) }
