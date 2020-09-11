package etcd

import (
	"crypto/sha1"
	"fmt"
)

// ETag computes the SHA1 hash for a given resource
func ETag(r interface{}) ([]byte, error) {
	bytes, err := marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to generate etag: %v", err)
	}

	return ETagFromBytes(bytes)
}

// ETagFromBytes computes the SHA1 hash for the given bytes
func ETagFromBytes(bytes []byte) ([]byte, error) {
	checksum := sha1.Sum(bytes)
	return checksum[:], nil
}
