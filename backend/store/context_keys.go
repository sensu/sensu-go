package store

// This file contains typed context keys that allocate zero bytes

type ContextKeyTimeoutT struct{}

// ContextKeyTimeout is used to specify a timeout as a time.Duration.
var ContextKeyTimeout = ContextKeyTimeoutT{}

func (ContextKeyTimeoutT) String() string {
	return "timeout"
}

func (ContextKeyTimeoutT) GoString() string {
	return "timeout"
}
