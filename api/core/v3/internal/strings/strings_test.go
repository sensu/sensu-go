package strings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFoundInArray(t *testing.T) {
	var array []string

	found := FoundInArray("Foo", []string{})
	assert.False(t, found)

	array = []string{"foo", "bar"}
	found = FoundInArray("Foo", array)
	assert.True(t, found)

	array = []string{"foo", "bar"}
	found = FoundInArray("FooBar", array)
	assert.False(t, found)

	array = []string{"foo", "bar"}
	found = FoundInArray("Foo ", array)
	assert.True(t, found)

	array = []string{"foo_bar"}
	found = FoundInArray("Foo_Bar", array)
	assert.True(t, found)

	array = []string{"foobar"}
	found = FoundInArray("Foo_Qux", array)
	assert.False(t, found)
}
