package agent

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBackendSelector(t *testing.T) {
	expected := []string{"a", "b", "c", "d", "e", "f", "g"}
	selector := &RandomBackendSelector{
		Backends: expected,
	}

	received := make([]string, len(selector.Backends))
	for i := 0; i < len(expected); i++ {
		received[i] = selector.Select()
	}

	sort.Strings(received)
	assert.EqualValues(t, expected, received)
}

func TestEmptyBackendSelector(t *testing.T) {
	selector := &RandomBackendSelector{}
	assert.Equal(t, "", selector.Select())
	assert.Equal(t, "", selector.Select())
}
