package etcd

import (
	"crypto/sha1"
	"fmt"
)

func ETag(r interface{}) ([]byte, error) {
	bytes, err := marshal(r)
	if err != nil {
		return nil, fmt.Errorf("failed to generate etag: %v", err)
	}
	checksum := sha1.Sum(bytes)
	return checksum[:], nil
}
