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
)

// Default error messages if not message is provided.
var standardErrorMessages = map[ErrCode]string{
	InternalErr:      "internal error occurred",
	InvalidArgument:  "invalid argument(s) received",
	NotFound:         "not found",
	AlreadyExistsErr: "resource already exists",
}

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
