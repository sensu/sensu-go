package stringutil

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

func TestFindInArray(t *testing.T) {
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

func TestRemove(t *testing.T) {
	array := []string{"bar", "qux"}
	array = Remove("bar", array)
	assert.Equal(t, []string{"qux"}, array)

	array = []string{}
	array = Remove("bar", array)
	assert.Equal(t, []string{}, array)
}

func TestIntersect(t *testing.T) {
	a, b := []string{"bar", "qux"}, []string{"foo", "bar"}
	intr := Intersect(a, b)
	assert.Equal(t, []string{"bar"}, intr)

	a, b = []string{}, []string{}
	intr = Intersect(a, b)
	assert.Equal(t, []string{}, intr)

	a, b = []string{"foo", "qux"}, []string{"baz", "bar"}
	intr = Intersect(a, b)
	assert.Equal(t, []string{}, intr)
}

func BenchmarkFoundInArray(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FoundInArray("foo!BAR   baz", []string{"foo", "foobar", "FOO BAR", "foo bar baz", "foobarbaz"})
	}
}
