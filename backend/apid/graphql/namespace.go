package graphql

import (
	"sort"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
)

const (
	// the maximum number of events & entities that will be read from the store
	// into memory; things likely won't work well if the upper bound is hit but
	// at least we aren't breaking the existing behaviour. Eventually this
	// interface will be deprecated in lieu of one that can be performant.
	maxSizeNamespaceListEvents = 25_000

	// When this number is exceeded the resolver will cease to count the total
	// number of entities. This should reduce the instances where we scan the
	// entire keyspace.
	maxCountNamespaceListEntities = 500

	// The maximum page size the list resolver will use when fetching entities
	// from the store.
	maxPageSizeNamespaceListEntities = 100
)

var _ schema.NamespaceFieldResolvers = (*namespaceImpl)(nil)

//
// Implement NamespaceFieldResolvers
//

type namespaceImpl struct {
	schema.MutatorAliases
	client       NamespaceClient
	eventClient  EventClient
	entityClient EntityClient
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

func listEntitiesOrdering(order schema.EntityListOrder) (string, bool) {
	switch order {
	case schema.EntityListOrders.ID:
		return corev2.EntitySortName, false
	default:
		return corev2.EntitySortName, true
	}
}

// Entities implements response to request for 'entities' field.
func (r *namespaceImpl) Entities(p schema.NamespaceEntitiesFieldResolverParams) (interface{}, error) {
	res := newOffsetContainer(p.Args.Offset, p.Args.Limit)
	ctx := store.NamespaceContext(p.Context, p.Source.(*corev2.Namespace).Name)

	ordering, desc := listEntitiesOrdering(p.Args.OrderBy)
	pred := &store.SelectionPredicate{
		Ordering:   ordering,
		Descending: desc,
		Limit:      int64(minInt(p.Args.Limit, maxPageSizeNamespaceListEntities)),
	}

	matches := 0
	records := make([]*corev2.Entity, 0, p.Args.Limit)

CONTINUE:
	queryResult, err := r.entityClient.ListEntities(ctx, pred)
	if err != nil {
		return res, err
	}

	// filter
	matchFn, err := filter.Compile(p.Args.Filters, EntityFilters(), corev2.EntityFields)
	if err != nil {
		return res, err
	}
	for i := range queryResult {
		if matchFn(queryResult[i]) {
			matches++
			if matches > p.Args.Offset && len(records) < p.Args.Limit {
				records = append(records, queryResult[i])
			}
		}
	}

	if pred.Continue != "" && matches < maxCountNamespaceListEntities {
		goto CONTINUE
	}

	res.Nodes = records
	res.PageInfo.totalCount = matches
	return res, nil
}

func listEventsOrdering(order schema.EventsListOrder) (string, bool) {
	switch order {
	case schema.EventsListOrders.ENTITY:
		return corev2.EventSortEntity, false
	case schema.EventsListOrders.ENTITY_DESC:
		return corev2.EventSortEntity, true
	case schema.EventsListOrders.LASTOK:
		return corev2.EventSortLastOk, true
	case schema.EventsListOrders.NEWEST:
		return corev2.EventSortTimestamp, true
	case schema.EventsListOrders.OLDEST:
		return corev2.EventSortTimestamp, false
	case schema.EventsListOrders.SEVERITY:
		return corev2.EventSortSeverity, false
	default:
		return corev2.EventSortLastOk, true
	}
}

func (r *namespaceImpl) eventsWithInStoreFiltering(p schema.NamespaceEventsFieldResolverParams) (interface{}, error) {
	res := newOffsetContainer(p.Args.Offset, p.Args.Limit)
	nsp := p.Source.(*corev2.Namespace)

	ctx := store.NamespaceContext(p.Context, nsp.Name)

	ordering, direction := listEventsOrdering(p.Args.OrderBy)
	pred := &store.SelectionPredicate{
		Limit:      int64(p.Args.Limit),
		Offset:     int64(p.Args.Offset),
		Ordering:   ordering,
		Descending: direction,
	}
	events, err := r.eventClient.ListEvents(ctx, pred)
	if err != nil {
		return res, err
	}
	// No predicate for all events in namespace
	totalResultCount, err := r.eventClient.CountEvents(ctx, nil)
	if err != nil {
		return res, err
	}

	res.Nodes = events
	res.PageInfo.totalCount = int(totalResultCount)
	return res, nil
}

// Events implements response to request for 'events' field.
func (r *namespaceImpl) Events(p schema.NamespaceEventsFieldResolverParams) (interface{}, error) {
	if r.eventClient.EventStoreSupportsFiltering(p.Context) {
		return r.eventsWithInStoreFiltering(p)
	}

	res := newOffsetContainer(p.Args.Offset, p.Args.Limit)
	nsp := p.Source.(*corev2.Namespace)

	// fetch
	ctx := store.NamespaceContext(p.Context, nsp.Name)
	results, err := listEvents(ctx, r.eventClient, "", maxSizeNamespaceListEvents)
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
