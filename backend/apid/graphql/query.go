package graphql

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/apid/graphql/suggest"
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
	factory      ClientFactory
}

// Viewer implements response to request for 'viewer' field.
func (r *queryImpl) Viewer(p graphql.ResolveParams) (interface{}, error) {
	return struct{}{}, nil
}

// Environment implements response to request for 'namespace' field.
func (r *queryImpl) Namespace(p schema.QueryNamespaceFieldResolverParams) (interface{}, error) {
	client := r.factory.NewWithContext(p.Context)
	res, err := client.FetchNamespace(p.Args.Name)
	return handleFetchResult(res, err)
}

// Event implements response to request for 'event' field.
func (r *queryImpl) Event(p schema.QueryEventFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	client := r.factory.NewWithContext(ctx)
	res, err := client.FetchEvent(p.Args.Entity, p.Args.Check)
	return handleFetchResult(res, err)
}

// Entity implements response to request for 'entity' field.
func (r *queryImpl) Entity(p schema.QueryEntityFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	client := r.factory.NewWithContext(ctx)
	res, err := client.FetchEntity(p.Args.Name)
	return handleFetchResult(res, err)
}

// Check implements response to request for 'check' field.
func (r *queryImpl) Check(p schema.QueryCheckFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	client := r.factory.NewWithContext(ctx)
	res, err := client.FetchCheck(p.Args.Name)
	return handleFetchResult(res, err)
}

// Handler implements a response to a request for the 'hander' field.
func (r *queryImpl) Handler(p schema.QueryHandlerFieldResolverParams) (interface{}, error) {
	ctx := contextWithNamespace(p.Context, p.Args.Namespace)
	client := r.factory.NewWithContext(ctx)
	res, err := client.FetchHandler(p.Args.Name)
	return handleFetchResult(res, err)
}

// Suggest implements a response to a request for the 'suggest' field.
func (r *queryImpl) Suggest(p schema.QuerySuggestFieldResolverParams) (interface{}, error) {
	results := map[string][]string{}
	results["values"] = []string{}

	ref, err := suggest.ParseRef(p.Args.Ref)
	if err != nil {
		return results, err
	}

	res := SuggestSchema.Lookup(ref)
	if res == nil {
		return results, errors.New("no configuration could be found for given ref")
	}

	field := res.LookupField(ref)
	if field == nil {
		return results, fmt.Errorf("could not find field for '%s'", ref.FieldPath)
	}

	client := r.factory.NewWithContext(p.Context)
	source := []v2.Resource{}

	err = client.List(res.URIPath(p.Args.Namespace), source, nil)
	if handleListErr(err) != nil {
		return results, err
	}

	q := strings.ToLower(p.Args.Q)
	set := utilstrings.OccurrenceSet{}
	for _, s := range source {
		for _, v := range field.Value(s, ref.FieldPath) {
			if strings.Contains(strings.ToLower(v), q) {
				set.Add(v)
			}
		}
	}

	values := set.Values()
	if p.Args.Order == schema.SuggestionOrders.FREQUENCY {
		sort.Strings(values)
		sort.SliceStable(values, func(i, j int) bool {
			return set.Get(values[i]) < set.Get(values[j])
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

func (impl *nodeImpl) ResolveType(i interface{}, _ graphql.ResolveTypeParams) *graphql.Type {
	resolver := impl.nodeResolver
	return resolver.FindType(i)
}
