package graphql

import (
	"context"

	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/graphql/tracing"
)

type InitHook func(*graphql.Service, ServiceConfig)

// InitHooks allow consumers to hook into the initialization of the service and
// mutate the schema. Useful for product variants.
var InitHooks = []InitHook{}

// ClientFactory instantiates new instances of the REST API client
type ClientFactory interface {
	NewWithContext(ctx context.Context) client.APIClient
}

// ServiceConfig describes values required to instantiate service.
type ServiceConfig struct {
	AssetClient       AssetClient
	CheckClient       CheckClient
	EntityClient      EntityClient
	EventClient       EventClient
	EventFilterClient EventFilterClient
	HandlerClient     HandlerClient
	MutatorClient     MutatorClient
	SilencedClient    SilencedClient
	NamespaceClient   NamespaceClient
	HookClient        HookClient
	UserClient        UserClient
	RBACClient        RBACClient
	GenericClient     GenericClient
}

// Service describes the Sensu GraphQL service capable of handling queries.
type Service struct {
	Target       *graphql.Service
	Config       ServiceConfig
	NodeRegister *relay.NodeRegister
}

// NewService instantiates new GraphQL service
func NewService(cfg ServiceConfig) (*Service, error) {
	svc := graphql.NewService()
	nodeResolver := newNodeResolver(cfg)
	wrapper := Service{
		Target:       svc,
		Config:       cfg,
		NodeRegister: nodeResolver.register,
	}

	// Register types
	schema.RegisterAsset(svc, &assetImpl{})
	schema.RegisterNamespace(svc, &namespaceImpl{client: cfg.NamespaceClient})
	schema.RegisterErrCode(svc)
	schema.RegisterEvent(svc, &eventImpl{})
	schema.RegisterEventsListOrder(svc)
	schema.RegisterIcon(svc)
	schema.RegisterJSON(svc, jsonImpl{})
	schema.RegisterKVPairString(svc, &schema.KVPairStringAliases{})
	schema.RegisterQuery(svc, &queryImpl{nodeResolver: nodeResolver, svc: cfg})
	schema.RegisterMutator(svc, &mutatorImpl{})
	schema.RegisterMutatorConnection(svc, &schema.MutatorConnectionAliases{})
	schema.RegisterMutatorListOrder(svc)
	schema.RegisterMutatorEdge(svc, &schema.MutatorEdgeAliases{})
	schema.RegisterMutedColour(svc)
	schema.RegisterNode(svc, &nodeImpl{nodeResolver})
	schema.RegisterNamespaced(svc, nil)
	schema.RegisterObjectMeta(svc, &objectMetaImpl{})
	schema.RegisterOffsetPageInfo(svc, &offsetPageInfoImpl{})
	schema.RegisterProxyRequests(svc, &schema.ProxyRequestsAliases{})
	schema.RegisterResource(svc, nil)
	schema.RegisterResolveEventPayload(svc, &schema.ResolveEventPayloadAliases{})
	schema.RegisterSchema(svc)
	schema.RegisterSilenceable(svc, nil)
	schema.RegisterSilenced(svc, &silencedImpl{client: cfg.CheckClient})
	schema.RegisterSilencedConnection(svc, &schema.SilencedConnectionAliases{})
	schema.RegisterSilencesListOrder(svc)
	schema.RegisterSubscriptionSet(svc, subscriptionSetImpl{})
	schema.RegisterSubscriptionSetOrder(svc)
	schema.RegisterSubscriptionOccurences(svc, &schema.SubscriptionOccurencesAliases{})
	schema.RegisterSuggestionOrder(svc)
	schema.RegisterSuggestionResultSet(svc, &schema.SuggestionResultSetAliases{})
	schema.RegisterUint(svc, unsignedIntegerImpl{})
	schema.RegisterViewer(svc, &viewerImpl{userClient: cfg.UserClient})

	// Register check types
	schema.RegisterCheck(svc, &checkImpl{})
	schema.RegisterCheckConfig(svc, &checkCfgImpl{})
	schema.RegisterCheckConfigConnection(svc, &schema.CheckConfigConnectionAliases{})
	schema.RegisterCheckHistory(svc, &checkHistoryImpl{})
	schema.RegisterCheckListOrder(svc)

	// Register entity types
	schema.RegisterEntity(svc, &entityImpl{})
	schema.RegisterEntityConnection(svc, &schema.EntityConnectionAliases{})
	schema.RegisterEntityListOrder(svc)
	schema.RegisterDeregistration(svc, &deregistrationImpl{})
	schema.RegisterNetwork(svc, &networkImpl{})
	schema.RegisterNetworkInterface(svc, &networkInterfaceImpl{})
	schema.RegisterSystem(svc, &systemImpl{})

	// Register event types
	schema.RegisterEvent(svc, &eventImpl{})
	schema.RegisterEventConnection(svc, &schema.EventConnectionAliases{})

	// Register event filter types
	schema.RegisterEventFilter(svc, &eventFilterImpl{})
	schema.RegisterEventFilterConnection(svc, &schema.EventFilterConnectionAliases{})
	schema.RegisterEventFilterAction(svc)
	schema.RegisterEventFilterListOrder(svc)

	// Register hook types
	schema.RegisterHook(svc, &hookImpl{})
	schema.RegisterHookConfig(svc, &hookCfgImpl{})
	schema.RegisterHookList(svc, &hookListImpl{})

	// Register handler types
	schema.RegisterHandler(svc, &handlerImpl{client: cfg.MutatorClient})
	schema.RegisterHandlerListOrder(svc)
	schema.RegisterHandlerConnection(svc, &schema.HandlerConnectionAliases{})
	schema.RegisterHandlerSocket(svc, &handlerSocketImpl{})

	// Register time window
	schema.RegisterTimeWindowDays(svc, &schema.TimeWindowDaysAliases{})
	schema.RegisterTimeWindowWhen(svc, &schema.TimeWindowWhenAliases{})
	schema.RegisterTimeWindowTimeRange(svc, &schema.TimeWindowTimeRangeAliases{})

	// Register RBAC types
	schema.RegisterClusterRole(svc, &schema.ClusterRoleAliases{})
	schema.RegisterClusterRoleBinding(svc, &schema.ClusterRoleBindingAliases{})
	schema.RegisterRole(svc, &schema.RoleAliases{})
	schema.RegisterRoleBinding(svc, &schema.RoleBindingAliases{})
	schema.RegisterRoleRef(svc, &schema.RoleRefAliases{})
	schema.RegisterRule(svc, &schema.RuleAliases{})
	schema.RegisterSubject(svc, &schema.SubjectAliases{})

	// Register user types
	schema.RegisterUser(svc, &userImpl{})

	// Register mutations
	schema.RegisterMutation(svc, &mutationsImpl{svc: cfg})
	schema.RegisterCheckConfigInputs(svc)
	schema.RegisterCreateCheckInput(svc)
	schema.RegisterCreateCheckPayload(svc, &checkMutationPayload{})
	schema.RegisterCreateSilenceInput(svc)
	schema.RegisterCreateSilencePayload(svc, &schema.CreateSilencePayloadAliases{})
	schema.RegisterDeleteRecordInput(svc)
	schema.RegisterDeleteRecordPayload(svc, &deleteRecordPayload{})
	schema.RegisterExecuteCheckInput(svc)
	schema.RegisterExecuteCheckPayload(svc, &schema.ExecuteCheckPayloadAliases{})
	schema.RegisterResolveEventInput(svc)
	schema.RegisterSilenceInputs(svc)
	schema.RegisterUpdateCheckInput(svc)
	schema.RegisterUpdateCheckPayload(svc, &checkMutationPayload{})
	schema.RegisterPutWrappedPayload(svc, &schema.PutWrappedPayloadAliases{})

	// Errors
	schema.RegisterStandardError(svc, stdErrImpl{})
	schema.RegisterError(svc, &errImpl{})

	// Run init hooks allowing consumers to extend service
	for _, hookFn := range InitHooks {
		hookFn(svc, cfg)
	}

	// Configure tracing
	tracer := tracing.NewPrometheusTracer()
	svc.RegisterMiddleware(tracer)

	err := svc.Regenerate()
	return &wrapper, err
}

// Do executes given query string and variables
func (svc *Service) Do(ctx context.Context, p graphql.QueryParams) *graphql.Result {
	// Instantiate loaders and lift them into the context
	qryCtx := contextWithLoaders(ctx, svc.Config)

	// Execute query inside context
	return svc.Target.Do(qryCtx, p)
}
