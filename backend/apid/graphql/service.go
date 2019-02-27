package graphql

import (
	"context"

	gql "github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/graphql"
)

// ClientFactory instantiates new instances of the REST API client
type ClientFactory interface {
	NewWithContext(ctx context.Context) client.APIClient
}

// ServiceConfig describes values required to instantiate service.
type ServiceConfig struct {
	ClientFactory ClientFactory
}

// Service describes the Sensu GraphQL service capable of handling queries.
type Service struct {
	target  *graphql.Service
	factory ClientFactory
}

// NewService instantiates new GraphQL service
func NewService(cfg ServiceConfig) (*Service, error) {
	svc := graphql.NewService()
	clientFactory := cfg.ClientFactory
	wrapper := Service{target: svc, factory: clientFactory}
	nodeResolver := newNodeResolver(clientFactory)

	// Register types
	schema.RegisterAsset(svc, &assetImpl{})
	schema.RegisterNamespace(svc, &namespaceImpl{factory: clientFactory})
	schema.RegisterErrCode(svc)
	schema.RegisterError(svc, nil)
	schema.RegisterEvent(svc, &eventImpl{})
	schema.RegisterEventsListOrder(svc)
	schema.RegisterHandler(svc, &handlerImpl{factory: clientFactory})
	schema.RegisterHandlerSocket(svc, &handlerSocketImpl{})
	schema.RegisterHasMetadata(svc, nil)
	schema.RegisterIcon(svc)
	schema.RegisterJSON(svc, jsonImpl{})
	schema.RegisterKVPairString(svc, &schema.KVPairStringAliases{})
	schema.RegisterQuery(svc, &queryImpl{nodeResolver: nodeResolver, factory: clientFactory})
	schema.RegisterMutator(svc, &mutatorImpl{})
	schema.RegisterMutedColour(svc)
	schema.RegisterNode(svc, &nodeImpl{nodeResolver})
	schema.RegisterNamespaced(svc, nil)
	schema.RegisterObjectMeta(svc, &objectMetaImpl{})
	schema.RegisterOffsetPageInfo(svc, &offsetPageInfoImpl{})
	schema.RegisterProxyRequests(svc, &schema.ProxyRequestsAliases{})
	schema.RegisterResolveEventPayload(svc, &schema.ResolveEventPayloadAliases{})
	schema.RegisterSchema(svc)
	schema.RegisterSilenced(svc, &silencedImpl{factory: clientFactory})
	schema.RegisterSilencedConnection(svc, &schema.SilencedConnectionAliases{})
	schema.RegisterStandardError(svc, stdErrImpl{})
	schema.RegisterSubscriptionSet(svc, subscriptionSetImpl{})
	schema.RegisterSubscriptionSetOrder(svc)
	schema.RegisterSubscriptionOccurences(svc, &schema.SubscriptionOccurencesAliases{})
	schema.RegisterSilencesListOrder(svc)
	schema.RegisterSilenceable(svc, nil)
	schema.RegisterUint(svc, unsignedIntegerImpl{})
	schema.RegisterViewer(svc, &viewerImpl{factory: clientFactory})

	// Register check types
	schema.RegisterCheck(svc, &checkImpl{factory: clientFactory})
	schema.RegisterCheckConfig(svc, &checkCfgImpl{factory: clientFactory})
	schema.RegisterCheckConfigConnection(svc, &schema.CheckConfigConnectionAliases{})
	schema.RegisterCheckHistory(svc, &checkHistoryImpl{})
	schema.RegisterCheckListOrder(svc)

	// Register entity types
	schema.RegisterEntity(svc, &entityImpl{factory: clientFactory})
	schema.RegisterEntityConnection(svc, &schema.EntityConnectionAliases{})
	schema.RegisterEntityListOrder(svc)
	schema.RegisterDeregistration(svc, &deregistrationImpl{})
	schema.RegisterNetwork(svc, &networkImpl{})
	schema.RegisterNetworkInterface(svc, &networkInterfaceImpl{})
	schema.RegisterSystem(svc, &systemImpl{})

	// Register event types
	schema.RegisterEvent(svc, &eventImpl{})
	schema.RegisterEventConnection(svc, &schema.EventConnectionAliases{})

	// Register hook types
	schema.RegisterHook(svc, &hookImpl{})
	schema.RegisterHookConfig(svc, &hookCfgImpl{})
	schema.RegisterHookList(svc, &hookListImpl{})

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
	schema.RegisterMutation(svc, &mutationsImpl{factory: clientFactory})
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

	err := svc.Regenerate()
	return &wrapper, err
}

// Do executes given query string and variables
func (svc *Service) Do(
	ctx context.Context,
	q string,
	vars map[string]interface{},
) *gql.Result {
	// Instantiate loaders and lift them into the context
	client := svc.factory.NewWithContext(ctx)
	qryCtx := contextWithLoaders(ctx, client)

	// Execute query inside context
	return svc.target.Do(qryCtx, q, vars)
}
