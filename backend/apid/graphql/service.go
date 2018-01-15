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

	// RegisterTypes
	schema.RegisterAsset(svc, &assetImpl{})
	schema.RegisterCheck(svc, &checkImpl{})
	schema.RegisterCheckConfig(svc, &checkCfgImpl{store: store})
	schema.RegisterCheckHistory(svc, &checkHistoryImpl{})
	schema.RegisterDeregistration(svc, &deregistrationImpl{})
	schema.RegisterEntity(svc, &entityImpl{})
	schema.RegisterHandler(svc, newHandlerImpl(store))
	schema.RegisterHandlerSocket(svc, &handlerSocketImpl{})
	schema.RegisterHook(svc, &hookImpl{})
	schema.RegisterHookConfig(svc, &hookCfgImpl{})
	schema.RegisterHookList(svc, &hookListImpl{})
	schema.RegisterQuery(svc, &queryImpl{store: store})
	schema.RegisterNetwork(svc, &networkImpl{})
	schema.RegisterNetworkInterface(svc, &networkInterfaceImpl{})
	schema.RegisterNode(svc, &nodeImpl{})
	schema.RegisterSchema(svc)

	err := svc.Regenerate()
	return svc, err
}
