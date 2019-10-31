package graphql

import (
	"errors"
	"sort"
	"strings"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	string_utils "github.com/sensu/sensu-go/util/strings"
)

var _ schema.NamespaceFieldResolvers = (*namespaceImpl)(nil)

//
// Implement NamespaceFieldResolvers
//

type namespaceImpl struct {
	client NamespaceClient
}

// ID implements response to request for 'id' field.
func (r *namespaceImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.NamespaceTranslator.EncodeToString(p.Context, p.Source), nil
}

// Name implements response to request for 'name' field.
func (r *namespaceImpl) Name(p graphql.ResolveParams) (string, error) {
	nsp := p.Source.(*corev2.Namespace)
	return nsp.Name, nil
}

// ColourID implements response to request for 'colourId' field.
// Experimental. Value is not persisted in any way at this time and is simply
// derived from the name.
func (r *namespaceImpl) ColourID(p graphql.ResolveParams) (schema.MutedColour, error) {
	nsp := p.Source.(*corev2.Namespace)
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
	nsp := p.Source.(*corev2.Namespace)

	// finds all records
	results, err := loadCheckConfigs(p.Context, nsp.Name)
	if err != nil {
		return res, err
	}

	// filter
	matches, err := filter.Compile(p.Args.Filters, CheckFilters(), corev2.CheckConfigFields)
	if err != nil {
		return res, err
	}
	filteredResults := make([]*corev2.CheckConfig, 0, len(results))
	for i := range results {
		if matches(results[i]) {
			filteredResults = append(filteredResults, results[i])
		}
	}

	// sort records
	sort.Sort(corev2.SortCheckConfigsByName(
		filteredResults,
		p.Args.OrderBy == schema.CheckListOrders.NAME,
	))

	// paginate
	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredResults))
	res.Nodes = filteredResults[l:h]
	res.PageInfo.totalCount = len(filteredResults)
	return res, nil
}

// EventFilters implements response to request for 'eventFilters' field.
func (r *namespaceImpl) EventFilters(p schema.NamespaceEventFiltersFieldResolverParams) (interface{}, error) {
	res := newOffsetContainer(p.Args.Offset, p.Args.Limit)
	nsp := p.Source.(*corev2.Namespace)

	// find all records
	results, err := loadEventFilters(p.Context, nsp.Name)
	if err != nil {
		return res, err
	}

	// filter
	matches, err := filter.Compile(p.Args.Filters, EventFilterFilters(), corev2.EventFilterFields)
	if err != nil {
		return res, err
	}
	filteredResults := make([]*corev2.EventFilter, 0, len(results))
	for i := range results {
		if matches(results[i]) {
			filteredResults = append(filteredResults, results[i])
		}
	}

	// sort records
	sort.Sort(corev2.SortEventFiltersByName(
		filteredResults,
		p.Args.OrderBy == schema.EventFilterListOrders.NAME,
	))

	// paginate
	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredResults))
	res.Nodes = filteredResults[l:h]
	res.PageInfo.totalCount = len(filteredResults)
	return res, nil
}

// Handlers implements response to request for 'handlers' field.
func (r *namespaceImpl) Handlers(p schema.NamespaceHandlersFieldResolverParams) (interface{}, error) {
	res := newOffsetContainer(p.Args.Offset, p.Args.Limit)
	nsp := p.Source.(*corev2.Namespace)

	// finds all records
	results, err := loadHandlers(p.Context, nsp.Name)
	if err != nil {
		return res, err
	}

	// filter
	matches, err := filter.Compile(p.Args.Filters, HandlerFilters(), corev2.HandlerFields)
	if err != nil {
		return res, err
	}
	filteredResults := make([]*corev2.Handler, 0, len(results))
	for i := range results {
		if matches(results[i]) {
			filteredResults = append(filteredResults, results[i])
		}
	}

	// sort
	sort.Sort(corev2.SortHandlersByName(
		filteredResults,
		p.Args.OrderBy == schema.HandlerListOrders.NAME,
	))

	// paginate
	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredResults))
	res.Nodes = filteredResults[l:h]
	res.PageInfo.totalCount = len(filteredResults)
	return res, nil
}

// Mutators implements response to request for 'mutators' field.
func (r *namespaceImpl) Mutators(p schema.NamespaceMutatorsFieldResolverParams) (interface{}, error) {
	res := newOffsetContainer(p.Args.Offset, p.Args.Limit)
	nsp := p.Source.(*corev2.Namespace)

	// finds all records
	results, err := loadMutators(p.Context, nsp.Name)
	if err != nil {
		return res, err
	}

	// filter
	matches, err := filter.Compile(p.Args.Filters, MutatorFilters(), corev2.MutatorFields)
	if err != nil {
		return res, err
	}
	filteredResults := make([]*corev2.Mutator, 0, len(results))
	for i := range results {
		if matches(results[i]) {
			filteredResults = append(filteredResults, results[i])
		}
	}

	// sort
	sort.Sort(corev2.SortMutatorsByName(
		filteredResults,
		p.Args.OrderBy == schema.MutatorListOrders.NAME,
	))

	// paginate
	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredResults))
	res.Nodes = filteredResults[l:h]
	res.PageInfo.totalCount = len(filteredResults)
	return res, nil
}

// Silences implements response to request for 'silences' field.
func (r *namespaceImpl) Silences(p schema.NamespaceSilencesFieldResolverParams) (interface{}, error) {
	res := newOffsetContainer(p.Args.Offset, p.Args.Limit)
	nsp := p.Source.(*corev2.Namespace)

	// fetch
	results, err := loadSilenceds(p.Context, nsp.Name)
	if err != nil {
		return res, err
	}

	// filter
	matches, err := filter.Compile(p.Args.Filters, SilenceFilters(), corev2.SilencedFields)
	if err != nil {
		return res, err
	}
	filteredResults := make([]*corev2.Silenced, 0, len(results))
	for i := range results {
		if matches(results[i]) {
			filteredResults = append(filteredResults, results[i])
		}
	}

	// sort records
	switch p.Args.OrderBy {
	case schema.SilencesListOrders.BEGIN_DESC:
		sort.Sort(sort.Reverse(corev2.SortSilencedByBegin(filteredResults)))
	case schema.SilencesListOrders.BEGIN:
		sort.Sort(corev2.SortSilencedByBegin(filteredResults))
	case schema.SilencesListOrders.ID_DESC:
		sort.Sort(sort.Reverse(corev2.SortSilencedByName(filteredResults)))
	case schema.SilencesListOrders.ID:
	default:
		sort.Sort(corev2.SortSilencedByName(filteredResults))
	}

	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredResults))
	res.Nodes = filteredResults[l:h]
	res.PageInfo.totalCount = len(filteredResults)
	return res, nil
}

// Entities implements response to request for 'entities' field.
func (r *namespaceImpl) Entities(p schema.NamespaceEntitiesFieldResolverParams) (interface{}, error) {
	res := newOffsetContainer(p.Args.Offset, p.Args.Limit)
	nsp := p.Source.(*corev2.Namespace)

	// fetch
	results, err := loadEntities(p.Context, nsp.Name)
	if err != nil {
		return res, err
	}

	// filter
	matches, err := filter.Compile(p.Args.Filters, EntityFilters(), corev2.EntityFields)
	if err != nil {
		return res, err
	}
	filteredResults := make([]*corev2.Entity, 0, len(results))
	for i := range results {
		if matches(results[i]) {
			filteredResults = append(filteredResults, results[i])
		}
	}

	// sort records
	switch p.Args.OrderBy {
	case schema.EntityListOrders.LASTSEEN:
		sort.Sort(corev2.SortEntitiesByLastSeen(filteredResults))
	default:
		sort.Sort(corev2.SortEntitiesByID(
			filteredResults,
			p.Args.OrderBy == schema.EntityListOrders.ID,
		))
	}

	// paginate
	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredResults))
	res.Nodes = filteredResults[l:h]
	res.PageInfo.totalCount = len(filteredResults)
	return res, nil
}

// Events implements response to request for 'events' field.
func (r *namespaceImpl) Events(p schema.NamespaceEventsFieldResolverParams) (interface{}, error) {
	res := newOffsetContainer(p.Args.Offset, p.Args.Limit)
	nsp := p.Source.(*corev2.Namespace)

	// fetch
	results, err := loadEvents(p.Context, nsp.Name)
	if err != nil {
		return res, err
	}

	// filter
	matches, err := filter.Compile(p.Args.Filters, EventFilters(), corev2.EventFields)
	if err != nil {
		return res, err
	}
	filteredResults := make([]*corev2.Event, 0, len(results))
	for i := range results {
		if matches(results[i]) {
			filteredResults = append(filteredResults, results[i])
		}
	}

	// sort records
	sortEvents(filteredResults, p.Args.OrderBy)

	// pagination
	l, h := clampSlice(p.Args.Offset, p.Args.Offset+p.Args.Limit, len(filteredResults))
	res.Nodes = filteredResults[l:h]
	res.PageInfo.totalCount = len(filteredResults)
	return res, nil
}

// Subscriptions implements response to request for 'subscriptions' field.
func (r *namespaceImpl) Subscriptions(p schema.NamespaceSubscriptionsFieldResolverParams) (interface{}, error) {
	set := string_utils.OccurrenceSet{}
	nsp := p.Source.(*corev2.Namespace)
	ctx := contextWithNamespace(p.Context, nsp.Name)

	// fetch
	entities, err := loadEntities(ctx, nsp.Name)
	if err != nil {
		return set, err
	}

	for i := range entities {
		entity := entities[i]
		newSet := occurrencesOfSubscriptions(entity)
		set.Merge(newSet)
	}

	checks, err := loadCheckConfigs(p.Context, nsp.Name)
	if err != nil {
		return set, err
	}
	for _, check := range checks {
		newSet := occurrencesOfSubscriptions(check)
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
	nsp := p.Source.(*corev2.Namespace)
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
