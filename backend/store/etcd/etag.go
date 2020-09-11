package etcd

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/hashstructure"
)

func ETag(v interface{}) (string, error) {
	hash, err := hashstructure.Hash(v, nil)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%q", strconv.FormatUint(hash, 10)), nil
}
