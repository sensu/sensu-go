//go:build integration
// +build integration

package etcd

import (
	"bytes"
	"context"
	"encoding/binary"
	"reflect"
	"strings"
	"testing"
)

func str(a uint64) string {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, a); err != nil {
		// Should never happen
		panic(err)
	}
	return buf.String()
}

// this is little more than a smoke test, it could use improvement
func TestSequence(t *testing.T) {
	e, cleanup := NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	seqA, err := Sequence(ctx, client, "/sensu.io/foobars/seq")
	if err != nil {
		t.Fatal(err)
	}

	seqB, err := Sequence(ctx, client, "/sensu.io/foobars/seq")
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

func TestSequences(t *testing.T) {
	e, cleanup := NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()

	got, err := Sequences(context.TODO(), client, "/sensu.io/foobars/seq", 5)
	if err != nil {
		t.Fatal(err)
	}

	if want := []string{str(1), str(2), str(3), str(4), str(5)}; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad sequences: got %v, want %v", got, want)
	}

	got, err = Sequences(context.TODO(), client, "/sensu.io/foobars/seq", 3)
	if err != nil {
		t.Fatal(err)
	}

	if want := []string{str(6), str(7), str(8)}; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad sequences: got %v, want %v", got, want)
	}

}
