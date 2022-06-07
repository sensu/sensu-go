package v2

import (
	"fmt"
)

// TODO(ccressent): Change these error structs to get rid of mentions of "Key"
// and replace with the appropriate concept (object name+namespace?)

// ErrAlreadyExists is returned when an object already exists.
type ErrAlreadyExists struct {
	Key string
}

func (e *ErrAlreadyExists) Error() string {
	return fmt.Sprintf("could not create the key %s", e.Key)
}

// ErrNotFound is returned when an object is not found in the store.
type ErrNotFound struct {
	Key string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("object %s not found", e.Key)
}

// ErrInternal is returned when something generally bad happened while
// interacting with the store. Other, more specific errors should be
// returned when appropriate.
//
// The backend will use ErrInternal to detect if an error is unrecoverable.
// It should only be used to signal that the underlying database is not
// functional
type ErrInternal struct {
	Message string
}

func (e *ErrInternal) Error() string {
	return fmt.Sprintf("internal error: %s", e.Message)
}
