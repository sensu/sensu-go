package graphql

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestSubscriptionSet(t *testing.T) {
	assert := assert.New(t)

	entity := types.FixtureEntity("test")
	entity.Subscriptions = []string{"one", "two"}

	// new
	setA := newSubscriptionSet(entity)
	assert.Equal(setA.size(), 2)
	assert.Contains(setA.entries(), "one")
	assert.Contains(setA.entries(), "two")

	// add
	setB := setA
	setB.add("three")
	setB.add("four")
	assert.Equal(setB.size(), 4)
	assert.Contains(setB.entries(), "three")
	assert.Contains(setB.entries(), "four")

	// merge
	setA.merge(setB)
	assert.Equal(setA.size(), 4)
	assert.Contains(setA.entries(), "three")
	assert.Contains(setA.entries(), "four")
}
