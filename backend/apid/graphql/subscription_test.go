package graphql

import (
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/util/strings"
	"github.com/stretchr/testify/assert"
)

func TestOccurrencesOfSubscription(t *testing.T) {
	assert := assert.New(t)

	entity := corev2.FixtureEntity("test")
	entity.Subscriptions = []string{"one", "two"}

	set := occurrencesOfSubscriptions(entity)
	assert.Equal(set.Size(), 2)
	assert.Equal(set.Get("one"), 1)
	assert.Equal(set.Get("two"), 1)
	assert.Equal(set.Get("three"), 0)
}

func TestSubscriptionSet(t *testing.T) {
	assert := assert.New(t)

	// new
	set := newSubscriptionSet(strings.NewOccurrenceSet("one", "one", "two"))
	assert.Equal(set.size(), 2)

	// sort alpha desc
	set.sortByAlpha(false)
	assert.EqualValues(set.values(), []string{"one", "two"})

	// sort alpha asc
	set.sortByAlpha(true)
	assert.EqualValues(set.values(), []string{"two", "one"})

	// sort by occurrence
	set.sortByOccurrence()
	assert.EqualValues(set.values(), []string{"one", "two"})

	// entries
	entries := set.entries()
	assert.Len(entries, 2)
	assert.EqualValues(entries, []subscriptionOccurrences{
		subscriptionOccurrences{"one", 2},
		subscriptionOccurrences{"two", 1},
	})
}
