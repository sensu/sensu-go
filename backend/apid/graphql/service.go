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
func NewService(cfg ServiceConfig) *graphql.Service {
	svc := graphql.NewService()
	store := cfg.Store

	// RegisterTypes
	schema.RegisterCheck(svc, &checkImpl{})
	schema.RegisterCheckConfig(svc, &checkCfgImpl{store: store})
	schema.RegisterCheckHistory(svc, &checkHistoryImpl{store: store})
	schema.RegisterQuery(svc, &queryImpl{store: store})

	return svc
}
