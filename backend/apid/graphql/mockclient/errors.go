package mockclient

import (
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/cli/client"
)

var (
	// InternalErr is used to help test internal errors
	InternalErr = client.APIError{Code: uint32(actions.InternalErr)}

	// InvalidArgument is used to help test invalid argument errors
	InvalidArgument = client.APIError{Code: uint32(actions.InvalidArgument)}

	// NotFound is used to help test not found errors
	NotFound = client.APIError{Code: uint32(actions.NotFound)}

	// AlreadyExistsErr is used to help test errors where resource already exists
	AlreadyExistsErr = client.APIError{Code: uint32(actions.AlreadyExistsErr)}
)
