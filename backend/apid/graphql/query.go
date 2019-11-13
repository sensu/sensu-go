package graphql

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/apid/graphql/suggest"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	utilstrings "github.com/sensu/sensu-go/util/strings"
)

var _ schema.QueryFieldResolvers = (*queryImpl)(nil)

//
// Implement QueryFieldResolvers
//

type queryImpl struct {
	nodeResolver *nodeResolver
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

	t, err := types.ResolveType(res.Group, res.Name)
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

	err = client.List(ctx, objs.Interface(), &store.SelectionPredicate{})
	if handleListErr(err) != nil {
		return results, err
	}

	q := strings.ToLower(p.Args.Q)
	set := utilstrings.OccurrenceSet{}
	for i := 0; i < objs.Elem().Len(); i++ {
		s := objs.Elem().Index(i).Interface().(corev2.Resource)
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
	nodeResolver *nodeResolver
}

func (impl *nodeImpl) ResolveType(i interface{}, p graphql.ResolveTypeParams) *graphql.Type {
	resolver := impl.nodeResolver
	return resolver.FindType(p.Context, i)
}
