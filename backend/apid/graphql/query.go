package graphql

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	apitools "github.com/sensu/sensu-api-tools"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/apid/graphql/suggest"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	utilstrings "github.com/sensu/sensu-go/util/strings"
)

var (
	// Defines the max amount of time we will allocate to fetching, filtering
	// and sorting suggestions
	suggestResolverTimeout = 850 * time.Millisecond
)

var _ schema.QueryFieldResolvers = (*queryImpl)(nil)

//
// Implement QueryFieldResolvers
//

type queryImpl struct {
	nodeResolver *relay.Resolver
	svc          ServiceConfig
}

// Viewer implements response to request for 'viewer' field.
func (r *queryImpl) Viewer(p graphql.ResolveParams) (interface{}, error) {
	return struct{}{}, nil
}

// Environment implements response to request for 'namespace' field.
func (r *queryImpl) Namespace(p schema.QueryNamespaceFieldResolverParams) (interface{}, error) {
	res, err := r.svc.NamespaceClient.FetchNamespace(p.Context, p.Args.Name)
	return handleFetchResult(res, err)
}

// Event implements response to request for 'event' field.
func (r *queryImpl) Event(p schema.QueryEventFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	res, err := r.svc.EventClient.FetchEvent(ctx, p.Args.Entity, p.Args.Check)
	return handleFetchResult(res, err)
}

// EventFilter implements response to request for 'eventFilter' field.
func (r *queryImpl) EventFilter(p schema.QueryEventFilterFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	res, err := r.svc.EventFilterClient.FetchEventFilter(ctx, p.Args.Name)
	return handleFetchResult(res, err)
}

// Entity implements response to request for 'entity' field.
func (r *queryImpl) Entity(p schema.QueryEntityFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	res, err := r.svc.EntityClient.FetchEntity(ctx, p.Args.Name)
	return handleFetchResult(res, err)
}

// Check implements response to request for 'check' field.
func (r *queryImpl) Check(p schema.QueryCheckFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	res, err := r.svc.CheckClient.FetchCheck(ctx, p.Args.Name)
	return handleFetchResult(res, err)
}

// Handler implements a response to a request for the 'hander' field.
func (r *queryImpl) Handler(p schema.QueryHandlerFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	res, err := r.svc.HandlerClient.FetchHandler(ctx, p.Args.Name)
	return handleFetchResult(res, err)
}

// Mutator implements a response to a request for the 'mutator' field.
func (r *queryImpl) Mutator(p schema.QueryMutatorFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	res, err := r.svc.MutatorClient.FetchMutator(ctx, p.Args.Name)
	return handleFetchResult(res, err)
}

// Suggest implements a response to a request for the 'suggest' field.
func (r *queryImpl) Suggest(p schema.QuerySuggestFieldResolverParams) (interface{}, error) {
	results := make(map[string]interface{}, 1)
	results["values"] = []string{}

	ref, err := suggest.ParseRef(p.Args.Ref)
	if err != nil {
		return results, err
	}

	res := SuggestSchema.Lookup(ref)
	if res == nil {
		return results, fmt.Errorf("no mapping found for '%s'", strings.Join([]string{ref.Group, ref.Name}, "/"))
	}

	field := res.LookupField(ref)
	if field == nil {
		return results, fmt.Errorf("could not find field for '%s'", ref.FieldPath)
	}

	t, err := apitools.Resolve(res.Group, res.Name)
	if err != nil {
		return results, err
	}

	// makes a slice from variable t and then gets the pointer to it.
	objT := reflect.MakeSlice(reflect.SliceOf(reflect.PtrTo(reflect.TypeOf(t).Elem())), 0, 0)
	objs := reflect.New(objT.Type())
	objs.Elem().Set(objT)

	client := r.svc.GenericClient
	// Don't need to check error, as type meta already successfully resolved
	_ = client.SetTypeMeta(corev2.TypeMeta{Type: res.Name, APIVersion: res.Group})

	ctx := store.NamespaceContext(p.Context, p.Args.Namespace)
	ctx, cancel := context.WithTimeout(ctx, suggestResolverTimeout)
	defer cancel()

	// CRUFT: entities can no longer be retrieved through the generic API
	// interface, to work around this we use the entity client.
	var entities []*corev2.Entity
	if res.Group == "core/v2" && res.Name == "entity" {
		entities, err = loadEntities(ctx, p.Args.Namespace)
		objs.Elem().Set(reflect.ValueOf(entities))
	} else {
		err = client.List(ctx, objs.Interface(), &store.SelectionPredicate{})
	}
	if handleListErr(err) != nil {
		return results, err
	}

	// if given one or more filters, configure a matcher
	var matches filter.Matcher = func(corev3.Resource) bool { return true }
	if len(p.Args.Filters) > 0 {
		matches, err = filter.Compile(p.Args.Filters, GlobalFilters, res.FilterFunc)
		if err != nil {
			return results, err
		}
	}

	q := strings.ToLower(p.Args.Q)
	set := utilstrings.OccurrenceSet{}
	for i := 0; i < objs.Elem().Len(); i++ {
		// IF the result set from the store was huge continue to check if we've
		// exceeded the deadline while we process the results. This feels a bit
		// crufty but may help avoid wasting a bunch of CPU time on a fairly
		// low priority process.
		if (i+1)%250 == 0 {
			if err := ctx.Err(); err != nil {
				break
			}
		}
		s := objs.Elem().Index(i).Interface().(corev3.Resource)
		if !matches(s) {
			continue
		}
		for _, v := range field.Value(s, ref.FieldPath) {
			if v != "" && strings.Contains(strings.ToLower(v), q) {
				set.Add(v)
			}
		}
	}

	values := set.Values()
	if p.Args.Order == schema.SuggestionOrders.FREQUENCY {
		sort.Strings(values)
		sort.SliceStable(values, func(i, j int) bool {
			return set.Get(values[i]) > set.Get(values[j])
		})
	} else if p.Args.Order == schema.SuggestionOrders.ALPHA_DESC {
		sort.Strings(values)
	} else if p.Args.Order == schema.SuggestionOrders.ALPHA_ASC {
		sort.Sort(sort.Reverse(sort.StringSlice(values)))
	}

	if len(values) > p.Args.Limit {
		values = values[:p.Args.Limit]
	}

	results["values"] = values
	return results, nil
}

// Versions implements response to request for 'versions' field.
func (r *queryImpl) Versions(p graphql.ResolveParams) (interface{}, error) {
	resp := r.svc.VersionController.GetVersion(p.Context)
	return resp, nil
}

// Health implements response to request for 'health' field.
func (r *queryImpl) Health(p graphql.ResolveParams) (interface{}, error) {
	return struct{}{}, nil
}

// Metrics implements response to request for 'metrics' field.
func (r *queryImpl) Metrics(p schema.QueryMetricsFieldResolverParams) (interface{}, error) {
	reg := r.svc.MetricGatherer
	mfs, err := reg.Gather()
	if err != nil {
		logger.WithError(err).Error("Query#metrics err while gathering metrics")
		if len(mfs) == 0 {
			return []interface{}{}, err
		}
	}
	mfsLen := len(mfs)
	if len(p.Args.Name) > 0 {
		mfsLen = len(p.Args.Name)
	}
	ret := make([]*dto.MetricFamily, 0, mfsLen)
	for _, mf := range mfs {
		if len(p.Args.Name) > 0 && !utilstrings.InArray(mf.GetName(), p.Args.Name) {
			continue
		}
		logger.WithField("fam", mf).Debug("has family")
		ret = append(ret, mf)
	}
	return ret, nil
}

// Node implements response to request for 'node' field.
func (r *queryImpl) Node(p schema.QueryNodeFieldResolverParams) (interface{}, error) {
	resolver := r.nodeResolver
	return resolver.Find(p.Context, p.Args.ID, p.Info)
}

// WrappedNode implements response to request for 'wrappedNode' field.
func (r *queryImpl) WrappedNode(p schema.QueryWrappedNodeFieldResolverParams) (interface{}, error) {
	resolver := r.nodeResolver
	res, err := resolver.Find(p.Context, p.Args.ID, p.Info)
	if rres, ok := res.(types.Resource); ok {
		return types.WrapResource(rres), err
	}
	return nil, err
}

//
// Implement Node interface
//

type nodeImpl struct {
	nodeResolver *relay.Resolver
}

func (impl *nodeImpl) ResolveType(i interface{}, p graphql.ResolveTypeParams) *graphql.Type {
	resolver := impl.nodeResolver
	return resolver.FindType(p.Context, i)
}
