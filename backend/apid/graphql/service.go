package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
)

// ServiceConfig describes values required to instantiate service.
type ServiceConfig struct {
	Store store.Store
	Bus   messaging.MessageBus
}

// NewService instantiates new GraphQL service
func NewService(cfg ServiceConfig) (*graphql.Service, error) {
	svc := graphql.NewService()
	store := cfg.Store

	// Register types
	schema.RegisterAsset(svc, &assetImpl{})
	schema.RegisterDeleteRecordInput(svc)
	schema.RegisterDeleteRecordPayload(svc, &deleteRecordPayload{})
	schema.RegisterHandler(svc, newHandlerImpl(store))
	schema.RegisterHandlerSocket(svc, &handlerSocketImpl{})
	schema.RegisterQuery(svc, &queryImpl{store: store})
	schema.RegisterMutation(svc, newMutationImpl(store))
	schema.RegisterMutator(svc, &mutatorImpl{})
	schema.RegisterNamespace(svc, &namespaceImpl{})
	schema.RegisterNode(svc, &nodeImpl{})
	schema.RegisterNamespaceInput(svc)
	schema.RegisterPageInfo(svc, &pageInfoImpl{})
	schema.RegisterSchema(svc)

	// Register check types
	schema.RegisterCheck(svc, &checkImpl{})
	schema.RegisterCheckConfig(svc, &checkCfgImpl{store: store})
	schema.RegisterCheckHistory(svc, &checkHistoryImpl{})
	schema.RegisterCheckConfigInputs(svc)
	schema.RegisterCreateCheckInput(svc)
	schema.RegisterCreateCheckPayload(svc, &checkMutationPayload{})
	schema.RegisterUpdateCheckInput(svc)
	schema.RegisterUpdateCheckPayload(svc, &checkMutationPayload{})

	// Register entity types
	schema.RegisterEntity(svc, &entityImpl{})
	schema.RegisterDeregistration(svc, &deregistrationImpl{})
	schema.RegisterNetwork(svc, &networkImpl{})
	schema.RegisterNetworkInterface(svc, &networkInterfaceImpl{})

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

	err := svc.Regenerate()
	return svc, err
}
