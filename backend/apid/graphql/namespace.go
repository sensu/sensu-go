package graphql

import (
	"errors"
	"sort"
	"strings"

	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/dynamic"
	string_utils "github.com/sensu/sensu-go/util/strings"
)

var _ schema.NamespaceFieldResolvers = (*namespaceImpl)(nil)

//
// Implement NamespaceFieldResolvers
//

type namespaceImpl struct {
	factory ClientFactory
}

// ID implements response to request for 'id' field.
func (r *namespaceImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.NamespaceTranslator.EncodeToString(p.Source), nil
}

// Name implements response to request for 'name' field.
func (r *namespaceImpl) Name(p graphql.ResolveParams) (string, error) {
	nsp := p.Source.(*types.Namespace)
	return nsp.Name, nil
}

// ColourID implements response to request for 'colourId' field.
// Experimental. Value is not persisted in any way at this time and is simply
// derived from the name.
func (r *namespaceImpl) ColourID(p graphql.ResolveParams) (schema.MutedColour, error) {
	nsp := p.Source.(*types.Namespace)
	num := nsp.Name[0] % 7
	switch num {
	case 0:
		return schema.MutedColours.BLUE, nil
	case 1:
		return schema.MutedColours.GRAY, nil
	case 2:
		return schema.MutedColours.GREEN, nil
	case 3:
		return schema.MutedColours.ORANGE, nil
	case 4:
		return schema.MutedColours.PINK, nil
	case 5:
		return schema.MutedColours.PURPLE, nil
	case 6:
		return schema.MutedColours.YELLOW, nil
	}
	return "", errors.New("exhausted list of colours")
}

// Checks implements response to request for 'checks' field.
func (r *namespaceImpl) Checks(p schema.NamespaceChecksFieldResolverParams) (interface{}, error) {
	res := newOffsetContainer(p.Args.Offset, p.Args.Limit)
	nsp := p.Source.(*types.Namespace)
	ctx := contextWithNamespace(p.Context, nsp.Name)

	client := r.factory.NewWithContext(ctx)
	records, err := client.ListChecks(nsp.Name)
	if err != nil {
		return res, err
	}

	// apply filters
	var filteredChecks []*types.CheckConfig
	filter := p.Args.Filter
	if len(filter) > 0 {
		filteredChecks = make([]*types.CheckConfig, 0, len(records))
		for _, record := range records {
			r := dynamic.Synthesize(record)
			matched, err := js.Evaluate(filter, r, nil)
			if err != nil {
				logger.WithError(err).Debug("unable to filter record")
				continue
			}
			if matched {
				filteredChecks = append(filteredChecks, &record)
			}
		}
	}

	// sort records
	sort.Sort(types.SortCheckConfigsByName(
		filteredChecks,
		p.Args.OrderBy == schema.CheckListOrders.NAME,
	))

	// paginate
	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredChecks))
	res.Nodes = filteredChecks[l:h]
	res.PageInfo.totalCount = len(filteredChecks)
	return res, nil
}

// Silences implements response to request for 'silences' field.
func (r *namespaceImpl) Silences(p schema.NamespaceSilencesFieldResolverParams) (interface{}, error) {
	res := newOffsetContainer(p.Args.Offset, p.Args.Limit)
	nsp := p.Source.(*types.Namespace)
	ctx := contextWithNamespace(p.Context, nsp.Name)

	// finds all records
	client := r.factory.NewWithContext(ctx)
	records, err := client.ListSilenceds(nsp.Name, "", "")
	if err != nil {
		return nil, err
	}

	// apply filters
	var filteredSilences []*types.Silenced
	filter := p.Args.Filter
	if len(filter) > 0 {
		filteredSilences = make([]*types.Silenced, 0, len(records))
		for _, r := range records {
			rec := dynamic.Synthesize(r)
			matched, err := js.Evaluate(filter, rec, nil)
			if err != nil {
				logger.WithError(err).Debug("unable to filter record")
				continue
			}
			if matched {
				filteredSilences = append(filteredSilences, &r)
			}
		}
	}

	// sort records
	switch p.Args.OrderBy {
	case schema.SilencesListOrders.BEGIN_DESC:
		sort.Sort(sort.Reverse(types.SortSilencedByBegin(filteredSilences)))
	case schema.SilencesListOrders.BEGIN:
		sort.Sort(types.SortSilencedByBegin(filteredSilences))
	case schema.SilencesListOrders.ID:
		sort.Sort(sort.Reverse(types.SortSilencedByName(filteredSilences)))
	case schema.SilencesListOrders.ID_DESC:
	default:
		sort.Sort(types.SortSilencedByName(filteredSilences))
	}

	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredSilences))
	res.Nodes = filteredSilences[l:h]
	res.PageInfo.totalCount = len(filteredSilences)
	return res, nil
}

// Entities implements response to request for 'entities' field.
func (r *namespaceImpl) Entities(p schema.NamespaceEntitiesFieldResolverParams) (interface{}, error) {
	res := newOffsetContainer(p.Args.Offset, p.Args.Limit)
	nsp := p.Source.(*types.Namespace)
	ctx := contextWithNamespace(p.Context, nsp.Name)

	client := r.factory.NewWithContext(ctx)
	records, err := client.ListEntities(nsp.Name)
	if err != nil {
		return nil, err
	}

	// apply filters
	filter := p.Args.Filter
	var filteredEntities []*types.Entity
	if len(filter) > 0 {
		filteredEntities = make([]*types.Entity, 0, len(filteredEntities))
		for i := range records {
			record := records[i]
			sr := dynamic.Synthesize(record.GetRedactedEntity())
			matched, err := js.Evaluate(filter, sr, nil)

			if err != nil {
				logger.WithError(err).Debug("unable to filter record")
				continue
			}
			if matched {
				filteredEntities = append(filteredEntities, &record)
			}
		}
	}

	// sort records
	switch p.Args.OrderBy {
	case schema.EntityListOrders.LASTSEEN:
		sort.Sort(types.SortEntitiesByLastSeen(filteredEntities))
	default:
		sort.Sort(types.SortEntitiesByID(
			filteredEntities,
			p.Args.OrderBy == schema.EntityListOrders.ID,
		))
	}

	// paginate
	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredEntities))
	res.Nodes = filteredEntities[l:h]
	res.PageInfo.totalCount = len(filteredEntities)
	return res, nil
}

// Events implements response to request for 'events' field.
func (r *namespaceImpl) Events(p schema.NamespaceEventsFieldResolverParams) (interface{}, error) {
	res := newOffsetContainer(p.Args.Offset, p.Args.Limit)
	nsp := p.Source.(*types.Namespace)
	ctx := contextWithNamespace(p.Context, nsp.Name)

	client := r.factory.NewWithContext(ctx)
	records, err := client.ListEvents(nsp.Name)
	if err != nil {
		return res, err
	}

	// apply filters
	filter := p.Args.Filter
	var filteredEvents []*types.Event
	if len(filter) > 0 {
		filteredEvents = make([]*types.Event, 0, len(records))
		for _, record := range records {
			r := dynamic.Synthesize(record)
			matched, err := js.Evaluate(filter, r, nil)
			if err != nil {
				logger.WithError(err).Debug("unable to filter record")
				continue
			}
			if matched {
				filteredEvents = append(filteredEvents, &record)
			}
		}
	}

	// sort records
	sortEvents(filteredEvents, p.Args.OrderBy)

	// pagination
	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredEvents))
	res.Nodes = filteredEvents[l:h]
	res.PageInfo.totalCount = len(filteredEvents)
	return res, nil
}

// CheckHistory implements response to request for 'checkHistory' field.
func (r *namespaceImpl) CheckHistory(p schema.NamespaceCheckHistoryFieldResolverParams) (interface{}, error) {
	nsp := p.Source.(*types.Namespace)
	ctx := contextWithNamespace(p.Context, nsp.Name)

	client := r.factory.NewWithContext(ctx)
	records, err := client.ListEvents(nsp.Name)
	if err != nil {
		return []types.CheckHistory{}, err
	}

	// Accumulate history
	history := []types.CheckHistory{}
	for _, record := range records {
		if record.Check == nil {
			continue
		}
		latest := types.CheckHistory{
			Executed: record.Check.Executed,
			Status:   record.Check.Status,
		}
		history = append(history, latest)
		history = append(history, record.Check.History...)
	}

	// Sort
	sort.Sort(types.ByExecuted(history))

	// Limit
	limit := clampInt(p.Args.Limit, 0, len(history))
	return history[0:limit], nil
}

// Subscriptions implements response to request for 'subscriptions' field.
func (r *namespaceImpl) Subscriptions(p schema.NamespaceSubscriptionsFieldResolverParams) (interface{}, error) {
	set := string_utils.OccurrenceSet{}
	nsp := p.Source.(*types.Namespace)
	ctx := contextWithNamespace(p.Context, nsp.Name)

	client := r.factory.NewWithContext(ctx)
	entities, err := client.ListEntities(nsp.Name)
	if err != nil {
		return set, err
	}
	for _, entity := range entities {
		newSet := occurrencesOfSubscriptions(&entity)
		set.Merge(newSet)
	}

	checks, err := client.ListChecks(nsp.Name)
	if err != nil {
		return set, err
	}
	for _, check := range checks {
		newSet := occurrencesOfSubscriptions(&check)
		set.Merge(newSet)
	}

	// If specified omit subscriptions prefix'd with 'entity:'
	if p.Args.OmitEntity {
		for _, subscription := range set.Values() {
			if strings.HasPrefix(subscription, "entity:") {
				set.Remove(subscription)
			}
		}
	}

	// Sort entries
	subscriptionSet := newSubscriptionSet(set)
	if p.Args.OrderBy == schema.SubscriptionSetOrders.ALPHA_DESC {
		subscriptionSet.sortByAlpha(false)
	} else if p.Args.OrderBy == schema.SubscriptionSetOrders.ALPHA_ASC {
		subscriptionSet.sortByAlpha(true)
	} else if p.Args.OrderBy == schema.SubscriptionSetOrders.OCCURRENCES {
		subscriptionSet.sortByOccurrence()
	}

	return subscriptionSet, nil
}

// IconID implements response to request for 'iconId' field.
// Experimental. Value is not persisted in any way at this time and is simply
// derived from the name.
func (r *namespaceImpl) IconID(p graphql.ResolveParams) (schema.Icon, error) {
	nsp := p.Source.(*types.Namespace)
	switch nsp.Name[0] % 11 {
	case 0:
		return schema.Icons.BRIEFCASE, nil
	case 1:
		return schema.Icons.DONUT, nil
	case 2:
		return schema.Icons.EMOTICON, nil
	case 3:
		return schema.Icons.ESPRESSO, nil
	case 4:
		return schema.Icons.EXPLORE, nil
	case 5:
		return schema.Icons.FIRE, nil
	case 6:
		return schema.Icons.HALFHEART, nil
	case 7:
		return schema.Icons.HEART, nil
	case 8:
		return schema.Icons.MUG, nil
	case 9:
		return schema.Icons.POLYGON, nil
	case 10:
		return schema.Icons.VISIBILITY, nil
	}
	return "", errors.New("exhausted list of icons")
}
