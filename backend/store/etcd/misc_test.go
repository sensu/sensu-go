package etcd

import (
	"reflect"
	"testing"
)

func testPartitionStrings(t *testing.T) {
	data := []string{"a", "a", "a", "a", "a", "a", "a", "a", "a"}
	want := [][]string{
		{"a", "a", "a", "a"},
		{"a", "a", "a", "a"},
		{"a"},
	}

	if got := partitionStrings(data, 4); !reflect.DeepEqual(got, want) {
		t.Fatalf("bad partitionStrings: got %v, want %v", got, want)
	}

	want = [][]string{{"a", "a", "a", "a", "a", "a", "a", "a", "a"}}

	if got := partitionStrings(data, 16); !reflect.DeepEqual(got, want) {
		t.Fatalf("bad partitionStrings: got %v, want %v", got, want)
	}
}
