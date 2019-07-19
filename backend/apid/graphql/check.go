package graphql

import (
	"time"

	"github.com/graphql-go/graphql"
	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/strings"
)

var _ schema.CheckFieldResolvers = (*checkImpl)(nil)
var _ schema.CheckConfigFieldResolvers = (*checkCfgImpl)(nil)
var _ schema.CheckHistoryFieldResolvers = (*checkHistoryImpl)(nil)

//
// Implement CheckConfigFieldResolvers
//

type checkCfgImpl struct {
	schema.CheckConfigAliases
}

// ID implements response to request for 'id' field.
func (r *checkCfgImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.CheckTranslator.EncodeToString(p.Source), nil
}

// ExtendedAttributes implements response to request for 'extendedAttributes' field.
func (*checkCfgImpl) ExtendedAttributes(p graphql.ResolveParams) (interface{}, error) {
	check := p.Source.(*types.CheckConfig)
	return wrapExtendedAttributes(check.ExtendedAttributes), nil
}

// Handlers implements response to request for 'handlers' field.
func (r *checkCfgImpl) Handlers(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*types.CheckConfig)
	results, err := loadHandlers(p.Context, src.Namespace)
	records := filterHandlers(results, func(obj *types.Handler) bool {
		return strings.FoundInArray(obj.Name, src.Handlers)
	})
	return records, err
}

// OutputMetricHandlers implements response to request for 'outputMetricHandlers' field.
func (r *checkCfgImpl) OutputMetricHandlers(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*types.CheckConfig)
	results, err := loadHandlers(p.Context, src.Namespace)
	records := filterHandlers(results, func(obj *types.Handler) bool {
		return strings.FoundInArray(obj.Name, src.OutputMetricHandlers)
	})
	return records, err
}

// IsSilenced implements response to request for 'isSilenced' field.
func (r *checkCfgImpl) IsSilenced(p graphql.ResolveParams) (bool, error) {
	src := p.Source.(*types.CheckConfig)
	now := time.Now().Unix()

	results, err := loadSilenceds(p.Context, src.Namespace)
	records := filterSilenceds(results, func(obj *types.Silenced) bool {
		if !obj.StartSilence(now) {
			return false
		}
		if (obj.Check == src.GetName() && (obj.Subscription == "" || obj.Subscription == "*")) ||
			((obj.Check == "" || obj.Check == "*") && strings.InArray(obj.Subscription, src.GetSubscriptions())) {
			return true
		}
		return false
	})
	return len(records) > 0, err
}

// Silences implements response to request for 'silences' field.
func (r *checkCfgImpl) Silences(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*types.CheckConfig)
	now := time.Now().Unix()

	results, err := loadSilenceds(p.Context, src.Namespace)
	records := filterSilenceds(results, func(obj *types.Silenced) bool {
		if !obj.StartSilence(now) {
			return false
		}
		if (obj.Check == src.GetName() && (obj.Subscription == "" || obj.Subscription == "*")) ||
			((obj.Check == "" || obj.Check == "*") && strings.InArray(obj.Subscription, src.GetSubscriptions())) {
			return true
		}
		return false
	})

	return records, err
}

// ToJSON implements response to request for 'toJSON' field.
func (r *checkCfgImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(v2.Resource)), nil
}

// RuntimeAssets implements response to request for 'runtimeAssets' field.
func (r *checkCfgImpl) RuntimeAssets(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*types.CheckConfig)
	records, err := loadAssets(p.Context, src.Namespace)
	results := filterAssets(records, func(obj *types.Asset) bool {
		return strings.FoundInArray(obj.Name, src.RuntimeAssets)
	})
	return results, err
}

// IsTypeOf is used to determine if a given value is associated with the Check type
func (r *checkCfgImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.CheckConfig)
	return ok
}

//
// Implement CheckFieldResolvers
//

type checkImpl struct {
	schema.CheckAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (r *checkImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.Check)
	return ok
}

// NodeID implements response to request for 'nodeId' field.
func (r *checkImpl) NodeID(p graphql.ResolveParams) (string, error) {
	check := p.Source.(*types.Check)
	config := types.CheckConfig{
		ObjectMeta: types.ObjectMeta{
			Namespace: check.Namespace,
			Name:      check.Name,
		},
	}
	return globalid.CheckTranslator.EncodeToString(&config), nil
}

// Executed implements response to request for 'executed' field.
func (r *checkImpl) Executed(p graphql.ResolveParams) (time.Time, error) {
	c := p.Source.(*types.Check)
	return time.Unix(c.Executed, 0), nil
}

// ExtendedAttributes implements response to request for 'extendedAttributes' field.
func (*checkImpl) ExtendedAttributes(p graphql.ResolveParams) (interface{}, error) {
	check := p.Source.(*types.Check)
	return wrapExtendedAttributes(check.ExtendedAttributes), nil
}

// LastOK implements response to request for 'lastOK' field.
func (r *checkImpl) LastOK(p graphql.ResolveParams) (*time.Time, error) {
	c := p.Source.(*types.Check)
	return convertTs(c.LastOK), nil
}

// Issued implements response to request for 'issued' field.
func (r *checkImpl) Issued(p graphql.ResolveParams) (time.Time, error) {
	c := p.Source.(*types.Check)
	return time.Unix(c.Issued, 0), nil
}

// History implements response to request for 'history' field.
func (r *checkImpl) History(p schema.CheckHistoryFieldResolverParams) (interface{}, error) {
	check := p.Source.(*types.Check)
	history := check.History

	length := clampInt(p.Args.First, 0, len(history))
	return history[0:length], nil
}

// Handlers implements response to request for 'handlers' field.
func (r *checkImpl) Handlers(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*types.Check)
	results, err := loadHandlers(p.Context, src.Namespace)
	records := filterHandlers(results, func(obj *types.Handler) bool {
		return strings.FoundInArray(obj.Name, src.Handlers)
	})
	return records, err
}

// IsSilenced implements response to request for 'isSilenced' field.
func (r *checkImpl) IsSilenced(p graphql.ResolveParams) (bool, error) {
	check := p.Source.(*types.Check)
	return len(check.Silenced) > 0, nil
}

// Silences implements response to request for 'silences' field.
func (r *checkImpl) Silences(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*types.Check)
	results, err := loadSilenceds(p.Context, src.Namespace)
	records := filterSilenceds(results, func(obj *types.Silenced) bool {
		return strings.FoundInArray(obj.Name, src.Silenced)
	})
	return records, err
}

// OutputMetricHandlers implements response to request for 'outputMetricHandlers' field.
func (r *checkImpl) OutputMetricHandlers(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*types.Check)
	results, err := loadHandlers(p.Context, src.Namespace)
	records := filterHandlers(results, func(obj *types.Handler) bool {
		return strings.FoundInArray(obj.Name, src.OutputMetricHandlers)
	})
	return records, err
}

// RuntimeAssets implements response to request for 'runtimeAssets' field.
func (r *checkImpl) RuntimeAssets(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*types.Check)
	records, err := loadAssets(p.Context, src.Namespace)
	results := filterAssets(records, func(obj *types.Asset) bool {
		return strings.FoundInArray(obj.Name, src.RuntimeAssets)
	})
	return results, err
}

// ToJSON implements response to request for 'toJSON' field.
func (r *checkImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(v2.Resource)), nil
}

//
// Implement CheckHistoryFieldResolvers
//

type checkHistoryImpl struct{}

// Status implements response to request for 'status' field.
func (r *checkHistoryImpl) Status(p graphql.ResolveParams) (interface{}, error) {
	h := p.Source.(types.CheckHistory)
	return h.Status, nil
}

// Executed implements response to request for 'executed' field.
func (r *checkHistoryImpl) Executed(p graphql.ResolveParams) (time.Time, error) {
	h := p.Source.(types.CheckHistory)
	return time.Unix(h.Executed, 0), nil
}
