package graphql

import (
	"errors"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	util_api "github.com/sensu/sensu-go/backend/apid/graphql/util/api"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
)

//
// Constants
//

var (
	// EntityCoreV3ConfigGlobalID can be used to produce a global id
	GlobalIDCoreV3EntityConfig = globalid.NewGenericTranslator(&corev3.EntityConfig{}, "")

	// EntityCoreV3StateGlobalID can be used to produce a global id
	GlobalIDCoreV3EntityState = globalid.NewGenericTranslator(&corev3.EntityState{}, "")
)

//
// EntityConfig
//

type corev3EntityConfigExtImpl struct {
	schema.CoreV3EntityConfigAliases
	client GenericClient
}

// ID implements response to request for 'id' field.
func (i *corev3EntityConfigExtImpl) ID(p graphql.ResolveParams) (string, error) {
	return GlobalIDCoreV3EntityConfig.EncodeToString(p.Context, p.Source), nil
}

// ToJSON implements response to request for 'toJSON' field.
func (i *corev3EntityConfigExtImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return util_api.WrapResource(p.Source), nil
}

// State implements response to request for 'state' field.
func (i *corev3EntityConfigExtImpl) State(p graphql.ResolveParams) (interface{}, error) {
	cfg := p.Source.(*corev3.EntityConfig)
	ctx := contextWithNamespace(p.Context, cfg.Metadata.Name)
	val := &corev3.V2ResourceProxy{Resource: &corev3.EntityState{}}
	err := i.client.Get(ctx, cfg.Metadata.Name, val)
	if errors.Is(err, &store.ErrNotFound{}) {
		return nil, nil
	}
	return val.Resource, err
}

type corev3EntityConfigImpl struct {
	schema.CoreV3EntityConfigAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*corev3EntityConfigImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev3.EntityConfig)
	return ok
}

//
// EntityState
//

type corev3EntityStateExtImpl struct {
	schema.CoreV3EntityStateAliases
	client GenericClient
}

// ID implements response to request for 'id' field.
func (*corev3EntityStateExtImpl) ID(p graphql.ResolveParams) (string, error) {
	return GlobalIDCoreV3EntityState.EncodeToString(p.Context, p.Source), nil
}

// ToJSON implements response to request for 'toJSON' field.
func (*corev3EntityStateExtImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return util_api.WrapResource(p.Source), nil
}

// State implements response to request for 'state' field.
func (i *corev3EntityStateExtImpl) Config(p graphql.ResolveParams) (interface{}, error) {
	state := p.Source.(*corev3.EntityState)
	ctx := contextWithNamespace(p.Context, state.Metadata.Name)
	val := &corev3.V2ResourceProxy{Resource: &corev3.EntityState{}}
	err := i.client.Get(ctx, state.Metadata.Name, val)
	if errors.Is(err, &store.ErrNotFound{}) {
		return nil, nil
	}
	return val.Resource, err
}

type corev3EntityStateImpl struct {
	schema.CoreV3EntityStateAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*corev3EntityStateImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev3.EntityState)
	return ok
}

func init() {
	globalid.RegisterTranslator(GlobalIDCoreV3EntityConfig)
	globalid.RegisterTranslator(GlobalIDCoreV3EntityState)
}
