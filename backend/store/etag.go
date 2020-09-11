package store

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/hashstructure"
)

// ETag returns a unique hash for the given interface
func ETag(v interface{}) (string, error) {
	hash, err := hashstructure.Hash(v, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%q", strconv.FormatUint(hash, 10)), nil
}

// ETagCondition represents a conditional request
type ETagCondition struct {
	IfMatch     string
	IfNoneMatch string
}
