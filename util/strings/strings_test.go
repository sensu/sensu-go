package strings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInArray(t *testing.T) {
	var item string
	var array []string

	found := InArray(item, array)
	assert.False(t, found, "if item and array are both empty, it should return false")

	item = "foo"
	found = InArray(item, array)
	assert.False(t, found, "if array is empty, it should return false")

	array = []string{"bar", "qux"}
	found = InArray(item, array)
	assert.False(t, found, "it should return false if the item isn't found in the array")

	array = append(array, "foo")
	found = InArray(item, array)
	assert.True(t, found, "it should return true if the item is found in the array")
}

func TestRemove(t *testing.T) {
	var array []string

	array = []string{"bar", "qux"}
	array = Remove("bar", array)
	assert.Equal(t, []string{"qux"}, array)

	array = []string{}
	array = Remove("bar", array)
	assert.Equal(t, []string{}, array)
}
