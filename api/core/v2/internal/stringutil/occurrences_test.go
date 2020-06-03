package stringutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOccurrences(t *testing.T) {
	assert := assert.New(t)

	// new
	setA := NewOccurrenceSet("one", "two")
	assert.Equal(setA.Size(), 2)
	assert.Contains(setA.Values(), "one")
	assert.Contains(setA.Values(), "two")

	// add
	setB := setA
	setB.Add("three")
	setB.Add("four")
	assert.Equal(setB.Size(), 4)
	assert.Contains(setB.Values(), "three")
	assert.Contains(setB.Values(), "four")

	// merge
	setA.Merge(setB)
	assert.Equal(setA.Size(), 4)
	assert.Contains(setA.Values(), "three")
	assert.Contains(setA.Values(), "four")

	// remove
	setA.Remove("four")
	assert.Equal(setA.Size(), 3)
	assert.NotContains(setA.Values(), "four")
}

func TestOccurrencesOf(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(OccurrencesOf("zero", []string{}), 0)
	assert.Equal(OccurrencesOf("zero", []string{"one", "one", "four"}), 0)
	assert.Equal(OccurrencesOf("one", []string{"one", "two", "three"}), 1)
	assert.Equal(OccurrencesOf("two", []string{"two", "two", "three"}), 2)
}
