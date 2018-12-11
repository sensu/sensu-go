package etcd

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3"
)

var initialItemKey []byte

func init() {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, uint64(0)); err != nil {
		// Should never happen
		panic(err)
	}
	initialItemKey = buf.Bytes()
}

// Sequence provides an incrementing sequence ID. The ID is a big-endian
// encoded uint64 value that starts at 0.
//
// If the key provided does not contain a sequence yet, one is created and
// incremented to 1. If the sequence already exists, then the next ID in the
// sequence will be returned.
//
// Because sequence IDs are big-endian encoded strings, they can be ordered
// with lexicographic sorting. This makes them a useful key for ordered
// collections.
//
// Sequence is not transactional, but is safe to use concurrently. If two
// clients are both using Sequence on the same key, they will race to be the
// one who updates the sequence. The loser of the race will execute the routine
// again.
func Sequence(kv clientv3.KV, key string) (result string, err error) {
	// Get the current key, or initialize it to be the first item key
	exists := clientv3.Compare(clientv3.Value(key), ">", string(initialItemKey))
	put := clientv3.OpPut(key, string(initialItemKey))
	get := clientv3.OpGet(key)

	resp, err := kv.Txn(context.Background()).If(exists).Then(get).Else(put, get).Commit()
	if err != nil {
		return "", fmt.Errorf("sequence error: %s", err)
	}

	respIdx := len(resp.Responses) - 1
	value := resp.Responses[respIdx].GetResponseRange().Kvs[0].Value

	// decode the key into an integer
	var n uint64
	if err := binary.Read(bytes.NewReader(value), binary.BigEndian, &n); err != nil {
		return "", fmt.Errorf("sequence error: %s", err)
	}

	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, n+1); err != nil {
		return "", fmt.Errorf("sequence error: %s", err)
	}

	notModified := clientv3.Compare(clientv3.Value(key), "=", string(value))
	put = clientv3.OpPut(key, buf.String())

	resp, err = kv.Txn(context.Background()).If(notModified).Then(put).Commit()
	if err != nil {
		return "", fmt.Errorf("sequence error: %s", err)
	}
	if !resp.Succeeded {
		return Sequence(kv, key)
	}

	return buf.String(), nil
}

// SequenceUint64 is like Sequence but returns its result as a uint64.
func SequenceUint64(kv clientv3.KV, key string) (uint64, error) {
	val, err := Sequence(kv, key)
	if err != nil {
		return 0, err
	}
	var result uint64
	if err := binary.Read(strings.NewReader(val), binary.BigEndian, &result); err != nil {
		return 0, fmt.Errorf("sequence error: %s", err)
	}
	return result, nil
}

// SetSequence sets a key to a particular sequence value.
func SetSequence(kv clientv3.KV, key string, value uint64) error {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, value); err != nil {
		return fmt.Errorf("couldn't set sequence: %s", err)
	}
	_, err := kv.Put(context.Background(), key, buf.String())
	return err
}
