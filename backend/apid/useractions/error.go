package useractions

import "fmt"

//
// Following defines error type w/ error codes. Helpful for
// determining which http status code to use in controllers and
// helpful for diving approriate return in GraphQL resolvers.
//

// ErrCode is an unsigned 32-bit integer used for defining error codes.
type ErrCode uint32

const (
	// InternalErr refers to an issue that occured in the underlying system. Eg.
	// if etcd was unreachable or something was implemented incorrectly.
	InternalErr ErrCode = 1

	// InvalidArgument refers to an issue with the arguments the user provided.
	// Eg. if arguments were of the wrong type or if the purposed changes cause
	// the record's validation to fail.
	InvalidArgument ErrCode = 2

	// NotFound means that the requested resource is unreachable or does not
	// exist.
	NotFound ErrCode = 3

	// AlreadyExistsErr means that a create operation failed because the given
	// resource already exists in the system.
	AlreadyExistsErr ErrCode = 4

	// PermissionDenied means that the viewer does not have permission to perform
	// the action they are attempting. Is not used when user is unauthenticated.
	// Eg. if the viewer is trying to list all events but doesn't not have the
	// approriate roles for the operation.
	PermissionDenied ErrCode = 5

	// Unauthenticated used when viewer is not authenticated but action requires
	// viewer to be authenticated.
	Unauthenticated ErrCode = 6
)

// Error describes an issue that ocurred while performing the action.
type Error struct {
	// Code refers to predefined codes that describe type of error that occurred.
	Code ErrCode
	// Message is a developer / operator friendly message briefly describing what
	// occurred.
	Message string
}

// Error method implements error interface
func (err Error) Error() string {
	return fmt.Sprintf("error: code = %d desc = %s", err.Code, err.Message)
}

// NewError returns a new Error given existing error and code.
func NewError(code ErrCode, err error) error {
	return Error{Code: code, Message: err.Error()}
}

// NewErrorf returns a new Error given message and code.
func NewErrorf(code ErrCode, f string, s ...interface{}) error {
	return Error{Code: code, Message: fmt.Sprintf(f, s...)}
}
