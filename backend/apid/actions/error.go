package actions

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
	InternalErr ErrCode = iota

	// InvalidArgument refers to an issue with the arguments the user provided.
	// Eg. if arguments were of the wrong type or if the purposed changes cause
	// the record's validation to fail.
	InvalidArgument

	// NotFound means that the requested resource is unreachable or does not
	// exist.
	NotFound

	// AlreadyExistsErr means that a create operation failed because the given
	// resource already exists in the system.
	AlreadyExistsErr

	// PermissionDenied means that the viewer does not have permission to perform
	// the action they are attempting. Is not used when user is unauthenticated.
	// Eg. if the viewer is trying to list all events but doesn't not have the
	// approriate roles for the operation.
	PermissionDenied

	// Unauthenticated used when viewer is not authenticated but action requires
	// viewer to be authenticated.
	Unauthenticated

	// PaymentRequired is used when the user tries to use a feature that's gated
	// behind a license.
	PaymentRequired

	// PreconditionFailed is used to indicate that a precondition header (e.g.
	// If-Match) is not matching the server side state
	PreconditionFailed

	// The deadline expired before the operation could complete. For operations
	// that change the state of the system, this error may be returned even if
	// the operation has completed successfully. For example, a successful
	// response from a server could have been delayed long
	DeadlineExceeded
)

// Default error messages if not message is provided.
var standardErrorMessages = map[ErrCode]string{
	InternalErr:        "internal error occurred",
	InvalidArgument:    "invalid argument(s) received",
	NotFound:           "not found",
	AlreadyExistsErr:   "resource already exists",
	PermissionDenied:   "unauthorized to perform action",
	Unauthenticated:    "unauthenticated",
	PaymentRequired:    "license required",
	PreconditionFailed: "precondition failed",
	DeadlineExceeded:   "deadline exceeded",
}

// Error describes an issue that ocurred while performing the action.
// TODO: This should likely be moved to the types package.
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
func NewError(code ErrCode, err error) Error {
	return Error{Code: code, Message: err.Error()}
}

// NewErrorf returns a new Error given message and code.
func NewErrorf(code ErrCode, s ...interface{}) Error {
	var f string
	if len(s) == 0 {
		f = standardErrorMessages[code]
	} else {
		f, s = s[0].(string), s[1:]
	}
	return Error{Code: code, Message: fmt.Sprintf(f, s...)}
}

// StatusFromError extracts code from the given error.
func StatusFromError(err error) (ErrCode, bool) {
	erro, ok := err.(Error)
	if !ok {
		return 0, false
	}

	return erro.Code, true
}
