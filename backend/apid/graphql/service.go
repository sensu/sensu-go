package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

// ServiceConfig describes values required to instantiate service.
type ServiceConfig struct {
	Store       store.Store
	Bus         messaging.MessageBus
	QueueGetter types.QueueGetter
}

// NewService instantiates new GraphQL service
func NewService(cfg ServiceConfig) (*graphql.Service, error) {
	svc := graphql.NewService()
	store := cfg.Store
	nodeResolver := newNodeResolver(store, cfg.QueueGetter)

	// Register types
	schema.RegisterAsset(svc, &assetImpl{})
	schema.RegisterEnvironment(svc, newEnvImpl(store, cfg.QueueGetter))
	schema.RegisterEnvironmentNode(svc, envNodeImpl{})
	schema.RegisterErrCode(svc)
	schema.RegisterError(svc, nil)
	schema.RegisterEvent(svc, &eventImpl{})
	schema.RegisterEventsListOrder(svc)
	schema.RegisterHandler(svc, newHandlerImpl(store))
	schema.RegisterHandlerSocket(svc, &handlerSocketImpl{})
	schema.RegisterIcon(svc)
	schema.RegisterJSON(svc, jsonImpl{})
	schema.RegisterQuery(svc, newQueryImpl(store, nodeResolver))
	schema.RegisterMutator(svc, &mutatorImpl{})
	schema.RegisterMutedColour(svc)
	schema.RegisterNamespace(svc, &namespaceImpl{})
	schema.RegisterNode(svc, &nodeImpl{nodeResolver})
	schema.RegisterNamespaceInput(svc)
	schema.RegisterOrganization(svc, newOrgImpl(store))
	schema.RegisterOffsetPageInfo(svc, &offsetPageInfoImpl{})
	schema.RegisterResolveEventPayload(svc, &schema.ResolveEventPayloadAliases{})
	schema.RegisterSchema(svc)
	schema.RegisterSilenced(svc, newSilencedImpl(store, cfg.QueueGetter))
	schema.RegisterSilencedConnection(svc, &schema.SilencedConnectionAliases{})
	schema.RegisterStandardError(svc, stdErrImpl{})
	schema.RegisterSubscriptionSet(svc, subscriptionSetImpl{})
	schema.RegisterSubscriptionSetOrder(svc)
	schema.RegisterSubscriptionOccurences(svc, &schema.SubscriptionOccurencesAliases{})
	schema.RegisterViewer(svc, newViewerImpl(store, cfg.QueueGetter, cfg.Bus))

	// Register check types
	schema.RegisterCheck(svc, &checkImpl{})
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
	schema.RegisterTimeWindowDays(svc, &timeWindowDaysImpl{})
	schema.RegisterTimeWindowWhen(svc, &timeWindowWhenImpl{})
	schema.RegisterTimeWindowTimeRange(svc, &timeWindowTimeRangeImpl{})

	// Register user types
	schema.RegisterRole(svc, &roleImpl{})
	schema.RegisterRule(svc, &ruleImpl{})
	schema.RegisterRuleResource(svc)
	schema.RegisterRulePermission(svc)
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
