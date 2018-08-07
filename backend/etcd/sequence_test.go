//+build integration

package etcd

import (
	"encoding/binary"
	"strings"
	"testing"
)

// this is little more than a smoke test, it could use improvement
func TestSequence(t *testing.T) {
	e, cleanup := NewTestEtcd(t)
	defer cleanup()

	client, err := e.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	seqA, err := Sequence(client, "/sensu.io/foobars/seq")
	if err != nil {
		t.Fatal(err)
	}

	seqB, err := Sequence(client, "/sensu.io/foobars/seq")
	if err != nil {
		t.Fatal(err)
	}

	var n1, n2 uint64
	if err := binary.Read(strings.NewReader(seqA), binary.BigEndian, &n1); err != nil {
		t.Fatal(err)
	}

	if got, want := n1, uint64(1); got != want {
		t.Errorf("bad id: got %d, want %d", got, want)
	}

	if err := binary.Read(strings.NewReader(seqB), binary.BigEndian, &n2); err != nil {
		t.Fatal(err)
	}

	if got, want := n2, uint64(2); got != want {
		t.Errorf("bad id: got %d, want %d", got, want)
	}
}
