package authorization

import "github.com/sensu/sensu-go/types"

// Authorizer determines whether a request is authorized using the RequestInfo
// passed. It returns true if the request should be authorized, along with any
// error encountered
type Authorizer interface {
	Authorize(reqInfo *types.RequestInfo) (bool, error)
}
