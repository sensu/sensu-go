package graphql

import (
	"errors"
	gostrings "strings"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/strings"
)

var _ schema.EventFilterFieldResolvers = (*eventFilterImpl)(nil)

//
// Implement EventFilterFieldResolvers
//

type eventFilterImpl struct {
	schema.EventFilterAliases
}

// ID implements response to request for 'id' field
func (*eventFilterImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.EventFilterTranslator.EncodeToString(p.Context, p.Source), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*eventFilterImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.EventFilter)
	return ok
}

// Action implements response to request for 'action' field.
func (*eventFilterImpl) Action(p graphql.ResolveParams) (schema.EventFilterAction, error) {
	src := p.Source.(*corev2.EventFilter)
	ret, ok := schema.EventFilterAction(gostrings.ToUpper(src.Action)), true
	if !ok {
		return ret, errors.New("unable to coerce value for field 'action'")
	}
	return ret, nil
}

// ToJSON implements response to request for 'toJSON' field.
func (*eventFilterImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(corev2.Resource)), nil
}

// RuntimeAssets implements response to request for 'runtimeAssets' field.
func (r *eventFilterImpl) RuntimeAssets(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*corev2.EventFilter)
	records, err := loadAssets(p.Context, src.Namespace)
	results := filterAssets(records, func(obj *corev2.Asset) bool {
		return strings.FoundInArray(obj.Name, src.RuntimeAssets)
	})
	return results, err
}
