package graphql

import (
	"errors"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

type corev3EntityConfigImpl struct {
	schema.CoreV3EntityConfigAliases
	client GenericClient
}

// ID implements response to request for 'id' field.
func (i *corev3EntityConfigImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.PipelineTranslator.EncodeToString(p.Context, p.Source), nil
}

// ToJSON implements response to request for 'toJSON' field.
func (i *corev3EntityConfigImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(types.Resource)), nil
}

// State implements response to request for 'state' field.
func (i *corev3EntityConfigImpl) State(p graphql.ResolveParams) (interface{}, error) {
	cfg := p.Source.(*corev3.EntityConfig)
	ctx := contextWithNamespace(p.Context, cfg.Metadata.Name)
	val := &corev3.V2ResourceProxy{Resource: &corev3.EntityState{}}
	err := i.client.Get(ctx, cfg.Metadata.Name, val)
	if errors.Is(err, &store.ErrNotFound{}) {
		return nil, nil
	}
	return val.Resource, err
}

// IsTypeOf is used to determine if a given value is associated with the type
func (i *corev3EntityConfigImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev3.EntityConfig)
	return ok
}

type corev3EntityStateImpl struct {
	schema.CoreV3EntityStateAliases
	client GenericClient
}

// ID implements response to request for 'id' field.
func (*corev3EntityStateImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.PipelineTranslator.EncodeToString(p.Context, p.Source), nil
}

// ToJSON implements response to request for 'toJSON' field.
func (*corev3EntityStateImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(types.Resource)), nil
}

// State implements response to request for 'state' field.
func (i *corev3EntityStateImpl) Config(p graphql.ResolveParams) (interface{}, error) {
	state := p.Source.(*corev3.EntityState)
	ctx := contextWithNamespace(p.Context, state.Metadata.Name)
	val := &corev3.V2ResourceProxy{Resource: &corev3.EntityState{}}
	err := i.client.Get(ctx, state.Metadata.Name, val)
	if errors.Is(err, &store.ErrNotFound{}) {
		return nil, nil
	}
	return val.Resource, err
}

// IsTypeOf is used to determine if a given value is associated with the type
func (i *corev3EntityStateImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev3.EntityState)
	return ok
}
