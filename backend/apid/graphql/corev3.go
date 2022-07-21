package graphql

import (
	"context"
	"errors"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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
	obj := p.Source.(*corev3.EntityConfig)
	val := corev3.V2ResourceProxy{Resource: &corev3.EntityState{}}
	err := getEntityComponent(p.Context, i.client, obj.Metadata, &val)
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
	obj := p.Source.(*corev3.EntityState)
	val := corev3.V2ResourceProxy{Resource: &corev3.EntityConfig{}}
	err := getEntityComponent(p.Context, i.client, obj.Metadata, &val)
	return val.Resource, err
}

func getEntityComponent(ctx context.Context, client GenericClient, meta *corev2.ObjectMeta, val corev2.Resource) error {
	wrapper := util_api.WrapResource(val)
	err := client.SetTypeMeta(wrapper.TypeMeta)
	if err != nil {
		return err
	}
	ctx = contextWithNamespace(ctx, meta.Name)
	if err := client.Get(ctx, meta.Name, val); err != nil && errors.Is(err, &store.ErrNotFound{}) {
		return nil
	}
	return err
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
