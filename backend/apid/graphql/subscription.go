package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
)

var _ schema.EnvironmentFieldResolvers = (*envImpl)(nil)

//
// Implement SubscriptionSetFieldResolvers
//

type subscriptionSetImpl struct{}

// Entries implements response to request for 'entries' field.
func (subscriptionSetImpl) Entries(p schema.SubscriptionSetEntriesFieldResolverParams) ([]string, error) {
	set := p.Source.(subscriptionSet)
	entries := set.entries()

	if p.Args.OrderBy == schema.SubscriptionSetOrders.ALPHA_DESC {
		sort.Strings(entries)
	} else if p.Args.OrderBy == schema.SubscriptionSetOrders.ALPHA_ASC {
		sort.Sort(sort.Reverse(sort.StringSlice(entries)))
	} else if p.Args.OrderBy == schema.SubscriptionSetOrders.FREQUENCY {
		sort.Sort(entries, func(i, j int) bool { set[entries[i]] < set[entries[j]] })
	}

	return entries, nil
}

// Size implements response to request for 'size' field.
func (subscriptionSetImpl) Size(p graphql.ResolveParams) (int, error) {
	set := p.Source.(subscriptionSet)
	return set.size(), nil
}

type subscribable interface {
	GetSubscriptions() []string
}

type subscriptionSet map[string]int

func newSubscriptionSet(record subscribable) {
	set := subscriptionSet{}
	for _, name := range record.GetSubscriptions() {
		set.add(name)
	}
	return set
}

func (set subscriptionSet) merge(b subscriptionSet) {
	for name, bCount := range b {
		aCount, _ := set[name]
		set[name] = aCount + bCount
	}
}

func (set subscriptionSet) add(name string) {
	num, _ := set[name]
	set[name] = num + 1
}

func (set subscriptionSet) size() int {
	return len(set)
}

func (set subscriptionSet) entries() []string {
	entries := []string{}
	for entry := range set {
		entries = append(entries, entry)
	}
	return entries
}
