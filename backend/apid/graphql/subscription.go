package graphql

import (
	"sort"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/util/strings"
)

var _ schema.SubscriptionSetFieldResolvers = (*subscriptionSetImpl)(nil)

//
// Implement SubscriptionSetFieldResolvers
//

type subscriptionSetImpl struct{}

// Entries implements response to request for 'entries' field.
func (subscriptionSetImpl) Entries(p schema.SubscriptionSetEntriesFieldResolverParams) (interface{}, error) {
	set := p.Source.(subscriptionSet)
	entries := set.entries()

	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(entries))
	return entries[l:h], nil
}

// Values implements response to request for 'values' field.
func (subscriptionSetImpl) Values(p schema.SubscriptionSetValuesFieldResolverParams) ([]string, error) {
	set := p.Source.(subscriptionSet)
	values := set.values()

	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(values))
	return values[l:h], nil
}

// Size implements response to request for 'size' field.
func (subscriptionSetImpl) Size(p graphql.ResolveParams) (int, error) {
	set := p.Source.(subscriptionSet)
	return set.size(), nil
}

// SubscriptionSet

type subscribable interface {
	GetSubscriptions() []string
}

func occurrencesOfSubscriptions(record subscribable) strings.OccurrenceSet {
	vls := record.GetSubscriptions()
	return strings.NewOccurrenceSet(vls...)
}

type subscriptionSet struct {
	occurrences strings.OccurrenceSet
	vls         []string
}

type subscriptionOccurrences struct {
	Subscription string
	Occurrences  int
}

func newSubscriptionSet(set strings.OccurrenceSet) subscriptionSet {
	return subscriptionSet{
		occurrences: set,
		vls:         set.Values(),
	}
}

func (set subscriptionSet) size() int {
	return len(set.vls)
}

func (set subscriptionSet) entries() []subscriptionOccurrences {
	occurrences := make([]subscriptionOccurrences, len(set.vls))
	for i, v := range set.vls {
		occurrences[i] = subscriptionOccurrences{v, set.occurrences.Get(v)}
	}
	return occurrences
}

func (set subscriptionSet) values() []string {
	return set.vls
}

func (set subscriptionSet) sortByAlpha(asc bool) {
	var vls sort.Interface = sort.StringSlice(set.vls)
	if asc {
		vls = sort.Reverse(vls)
	}
	sort.Sort(vls)
}

func (set subscriptionSet) sortByOccurrence() {
	sort.Slice(set.vls, func(i, j int) bool {
		a := set.occurrences.Get(set.vls[i])
		b := set.occurrences.Get(set.vls[j])
		return a > b
	})
}
