package graphql

import (
	"context"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

type ClientFactory interface {
	NewWithContext(ctx context.Context) client.APIClient
}

// ServiceConfig describes values required to instantiate service.
type ServiceConfig struct {
	Store         store.Store
	Bus           messaging.MessageBus
	QueueGetter   types.QueueGetter
	ClientFactory ClientFactory
}

// NewService instantiates new GraphQL service
func NewService(cfg ServiceConfig) (*graphql.Service, error) {
	svc := graphql.NewService()
	store := cfg.Store
	clientFactory := cfg.ClientFactory
	nodeResolver := newNodeResolver(store, cfg.QueueGetter)

	// Register types
	schema.RegisterAsset(svc, &assetImpl{})
	schema.RegisterNamespace(svc, &namespaceImpl{factory: clientFactory})
	schema.RegisterErrCode(svc)
	schema.RegisterError(svc, nil)
	schema.RegisterEvent(svc, &eventImpl{})
	schema.RegisterEventsListOrder(svc)
	schema.RegisterHandler(svc, newHandlerImpl(store))
	schema.RegisterHandlerSocket(svc, &handlerSocketImpl{})
	schema.RegisterIcon(svc)
	schema.RegisterJSON(svc, jsonImpl{})
	schema.RegisterQuery(svc, newQueryImpl(store, nodeResolver, cfg.QueueGetter))
	schema.RegisterMutator(svc, &mutatorImpl{})
	schema.RegisterMutedColour(svc)
	schema.RegisterNode(svc, &nodeImpl{nodeResolver})
	schema.RegisterNamespaced(svc, nil)
	schema.RegisterOffsetPageInfo(svc, &offsetPageInfoImpl{})
	schema.RegisterProxyRequests(svc, &schema.ProxyRequestsAliases{})
	schema.RegisterResolveEventPayload(svc, &schema.ResolveEventPayloadAliases{})
	schema.RegisterSchema(svc)
	schema.RegisterSilenced(svc, newSilencedImpl(store, cfg.QueueGetter))
	schema.RegisterSilencedConnection(svc, &schema.SilencedConnectionAliases{})
	schema.RegisterStandardError(svc, stdErrImpl{})
	schema.RegisterSubscriptionSet(svc, subscriptionSetImpl{})
	schema.RegisterSubscriptionSetOrder(svc)
	schema.RegisterSubscriptionOccurences(svc, &schema.SubscriptionOccurencesAliases{})
	schema.RegisterSilencesListOrder(svc)
	schema.RegisterViewer(svc, newViewerImpl(store))

	// Register check types
	schema.RegisterCheck(svc, newCheckImpl(store))
	schema.RegisterCheckConfig(svc, newCheckCfgImpl(store))
	schema.RegisterCheckConfigConnection(svc, &schema.CheckConfigConnectionAliases{})
	schema.RegisterCheckHistory(svc, &checkHistoryImpl{})
	schema.RegisterCheckListOrder(svc)

	// Register entity types
	schema.RegisterEntity(svc, newEntityImpl(store))
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
	schema.RegisterMutation(svc, newMutationImpl(store, cfg.QueueGetter, cfg.Bus))
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
	return svc, err
}
