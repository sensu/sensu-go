package graphqlschema

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql/globalid"
	"github.com/sensu/sensu-go/graphql/relay"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
)

var errMock = errors.New("v spoopy error message")

type setContextFn func(context.Context) context.Context

func newParams(source interface{}, fns ...setContextFn) graphql.ResolveParams {
	params := graphql.ResolveParams{
		Source:  source,
		Context: context.TODO(),
		Args:    map[string]interface{}{},
	}

	paramsContextDefaults := []setContextFn{
		contextWithOrgEnv("", ""),
		contextWithFullAccess,
	}
	fns = append(paramsContextDefaults, fns...)
	for _, fn := range fns {
		params.Context = fn(params.Context)
	}

	return params
}

func contextWithOrgEnv(org, env string) setContextFn {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, types.EnvironmentKey, env)
		ctx = context.WithValue(ctx, types.OrganizationKey, org)
		return ctx
	}
}

func contextWithFullAccess(ctx context.Context) context.Context {
	userRules := []types.Rule{*types.FixtureRule("*", "*")}
	actor := authorization.Actor{Name: "sensu", Rules: userRules}
	return context.WithValue(ctx, types.AuthorizationActorKey, actor)
}

func contextWithNoAccess(ctx context.Context) context.Context {
	return context.WithValue(
		ctx,
		types.AuthorizationActorKey,
		authorization.Actor{},
	)
}

func contextWithStore(store store.Store) setContextFn {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, types.StoreKey, store)
	}
}

func fetchResolver(T *graphql.Object, fName string) graphql.FieldResolveFn {
	fields := T.Fields()
	return fields[fName].Resolve
}

// resolverSuite is written to work along side testify's suite package, exposing
// methods to make testing resolvers easy.
//
// Example:
//
//   type UserTypeNameField struct {
//     suite.Suite
//     resolverSuite
//   }
//
//   func (t *UserTypeNameField) TestFullAccess() {
//     user := User{Username: "FrankWest"}
//     result, err := t.runResolver("User.Username", user)
//
//     t.Equal("frank-west", result)
//     t.NoError(err)
//   }
//
//   func (t *UserTypeNameField) TestNoAccess() {
//     user   := User{Username: "FrankWest"}
//     params := t.newParams(user, contextWithNoAccess)
//     result, err := t.runResolver("User.Username", params)
//
//     t.Equal("frank-west", result)
//     t.NoError(err)
//   }
//
type resolverSuite struct {
	_store mockstore.MockStore
}

// newParams inits new graphql resolver params, lifts store into context and
// returns params.
//
// Example:
//
//   func (t *UserTypeNameField) TestNoAccess() {
//     user   := User{Username: "FrankWest"}
//     params := t.newParams(user, contextWithNoAccess)
//     result, err := t.runResolver("User.Username", params)
//
//     t.Equal("frank-west", result)
//     t.NoError(err)
//   }
//
func (t *resolverSuite) newParams(source interface{}, fns ...setContextFn) graphql.ResolveParams {
	fns = append([]setContextFn{contextWithStore(t.store())}, fns...)
	params := newParams(source, fns...)
	return params
}

// runResolver given field and source|params, runs resolver and returns results.
//
// Example:
//
//   func (t *UserTypeNameField) TestWithSource() {
//     source := User{Username: "FrankWest"}
//     result, err := t.runResolver("User.Username", source)
//     t.Equal("frank-west", result)
//   }
//
//   func (t *UserTypeNameField) TestWithParams() {
//     params := t.newParams(nil)
//     result, err := t.runResolver("User.Username", params)
//     t.Equal("", result)
//   }
//
func (t *resolverSuite) runResolver(typeDotField string, paramsOrSource interface{}) (interface{}, error) {
	// Instantiate params if we ever just given a source
	var params graphql.ResolveParams
	switch v := paramsOrSource.(type) {
	case graphql.ResolveParams:
		params = v
	default:
		params = t.newParams(paramsOrSource)
	}

	// variable should be in the format ObjecType.Field.
	// (eg. Viewer.entities, User.Username, Check.Name, ...)
	fieldComponents := strings.Split(typeDotField, ".")

	// Instantiate schema and find given type
	schema := Schema()
	types := schema.TypeMap()
	typeName := fieldComponents[0]
	objectType := types[typeName].(*graphql.Object)

	// Fetch field from object type
	typeFields := objectType.Fields()
	fieldName := fieldComponents[1]
	resolver := typeFields[fieldName].Resolve

	// Run resolver
	result, err := resolver(params)
	return result, err
}

// store returns initiailized mockstore for easily mocking etcd requests
func (t *resolverSuite) store() *mockstore.MockStore {
	return &t._store
}

// nodeSuite is written to work in concert with testify's suite package,
// exposing methods to make testing node resolvers easy.
//
// Example:
//
//   type UserNode struct {
//     suite.Suite
//     resolverSuite
//   }
//
//   func (t *UserNode) SetupTest() {
//     t.setNodeResolver(userNodeResolver)
//   }
//
//   func (t *UserTypeNameField) TestFullAccess() {
//     user := User{Username: "FrankWest"}
//     result, err := t.runResolver(user)
//
//     t.Equal("frank-west", result)
//     t.NoError(err)
//   }
//
type nodeSuite struct {
	_store        mockstore.MockStore
	_nodeResolver *relay.NodeResolver
}

// newParams inits new node resolver params, lifts store into context and
// returns params.
//
// Example:
//
//   func (t *UserTypeNameField) TestNoAccess() {
//     user   := User{Username: "FrankWest"}
//     params := t.newParams(user, contextWithNoAccess)
//     result, err := t.runResolver(myResolver, params)
//
//     t.Equal(user, result)
//     t.NoError(err)
//   }
//
func (t *nodeSuite) newParams(source interface{}, fns ...setContextFn) relay.NodeResolverParams {
	// Encode ID
	translator := t.nodeResolver().Translator
	idComponents := translator.Encode(source).(globalid.StandardComponents)

	// Configure params
	params := relay.NodeResolverParams{
		IDComponents: translator.Decode(idComponents),
		Context:      context.TODO(),
	}

	// Lift defaults into context and apply given setters
	contextDefaults := []setContextFn{
		contextWithOrgEnv("", ""),
		contextWithFullAccess,
		contextWithStore(t.store()),
	}
	for _, fn := range append(contextDefaults, fns...) {
		params.Context = fn(params.Context)
	}

	return params
}

// runResolver given NodeResolver and source|params, runs resolver and returns
// results.
//
// Example:
//
//   func (t *UserTypeNameField) TestWithSource() {
//     source := User{Username: "FrankWest"}
//     result, err := t.runResolver(source)
//     t.Equal(source, result)
//     t.NoError(err)
//   }
//
//   func (t *UserTypeNameField) TestWithParams() {
//     user := User{Username: "FrankWest"}
//     params := t.newParams(user, contextWithNoAccess)
//     result, err := t.runResolver(params)
//     t.Nil(result)
//     t.NoError(err)
//   }
//
func (t *nodeSuite) runResolver(paramsOrSource interface{}) (interface{}, error) {
	// Instantiate params if we ever just given a source
	var params relay.NodeResolverParams
	switch v := paramsOrSource.(type) {
	case relay.NodeResolverParams:
		params = v
	default:
		params = t.newParams(paramsOrSource)
	}

	result, err := t.nodeResolver().Resolve(params)
	return result, err
}

// store returns initiailized mockstore for easily mocking etcd requests
func (t *nodeSuite) store() *mockstore.MockStore {
	return &t._store
}

// nodeResolver returns configured node resolver
func (t *nodeSuite) nodeResolver() *relay.NodeResolver {
	return t._nodeResolver
}

// setNodeResolver returns configured node resolver
func (t *nodeSuite) setNodeResolver(r *relay.NodeResolver) {
	t._nodeResolver = r
}

func runSuites(t *testing.T, suites ...suite.TestingSuite) {
	for _, s := range suites {
		t.Run(reflect.TypeOf(s).Elem().Name(), func(t *testing.T) {
			suite.Run(t, s)
		})
	}
}
