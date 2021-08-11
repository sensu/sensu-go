package etcd

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"

	"go.etcd.io/etcd/client/v3"
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
func Sequence(ctx context.Context, kv clientv3.KV, key string) (result string, err error) {
	// Get the current key, or initialize it to be the first item key
	seqs, err := Sequences(ctx, kv, key, 1)
	if err != nil {
		return "", err
	}
	if len(seqs) == 0 {
		return "", errors.New("sequences returned no results")
	}
	return seqs[0], nil
}

// Sequences is like Sequence, but returns a slice of sequences whose length is
// equal to values.
func Sequences(ctx context.Context, kv clientv3.KV, key string, values int) ([]string, error) {
	// Get the current key, or initialize it to be the first item key
	if values == 0 {
		return nil, nil
	}
	if values < 0 {
		return nil, fmt.Errorf("negative values requested")
	}
	exists := clientv3.Compare(clientv3.Version(key), ">", 0)
	put := clientv3.OpPut(key, string(initialItemKey))
	get := clientv3.OpGet(key)

	resp, err := kv.Txn(context.Background()).If(exists).Then(get).Else(put, get).Commit()
	if err != nil {
		return nil, fmt.Errorf("sequence error: %s", err)
	}

	respIdx := len(resp.Responses) - 1
	value := resp.Responses[respIdx].GetResponseRange().Kvs[0].Value

	// decode the key into an integer
	var n uint64
	if err := binary.Read(bytes.NewReader(value), binary.BigEndian, &n); err != nil {
		return nil, fmt.Errorf("sequence error reading sequence: %s", err)
	}

	var buf *bytes.Buffer
	result := make([]string, values)
	for i := range result {
		buf = new(bytes.Buffer)
		if err := binary.Write(buf, binary.BigEndian, n+1); err != nil {
			return nil, fmt.Errorf("sequence error writing sequence: %s", err)
		}
		result[i] = buf.String()
		n++
	}

	notModified := clientv3.Compare(clientv3.Value(key), "=", string(value))
	put = clientv3.OpPut(key, buf.String())

	resp, err = kv.Txn(context.Background()).If(notModified).Then(put).Commit()
	if err != nil {
		return nil, fmt.Errorf("sequence error: %s", err)
	}
	if !resp.Succeeded {
		return Sequences(ctx, kv, key, values)
	}

	return result, nil
}
