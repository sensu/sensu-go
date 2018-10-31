package graphql

import (
	"context"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/strings"
)

var _ schema.CheckFieldResolvers = (*checkImpl)(nil)
var _ schema.CheckConfigFieldResolvers = (*checkCfgImpl)(nil)
var _ schema.CheckHistoryFieldResolvers = (*checkHistoryImpl)(nil)

type namedCheck interface {
	GetName() string
	GetSubscriptions() []string
}

type silenceableCheck interface {
	namedCheck
	GetSilenced() []string
}

//
// Implement CheckConfigFieldResolvers
//

type checkCfgImpl struct {
	schema.CheckConfigAliases
	handlerCtrl    actions.HandlerController
	silenceQuerier silenceQuerier
}

func newCheckCfgImpl(store store.Store) *checkCfgImpl {
	handlerCtrl := actions.NewHandlerController(store)
	silenceCtrl := actions.NewSilencedController(store)

	return &checkCfgImpl{
		handlerCtrl:    handlerCtrl,
		silenceQuerier: silenceCtrl,
	}
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
	check := p.Source.(*types.CheckConfig)
	ctx := types.SetContextFromResource(p.Context, check)
	return fetchHandlersWithNames(ctx, r.handlerCtrl, check.Handlers)
}

// OutputMetricHandlers implements response to request for 'outputMetricHandlers' field.
func (r *checkCfgImpl) OutputMetricHandlers(p graphql.ResolveParams) (interface{}, error) {
	check := p.Source.(*types.CheckConfig)
	ctx := types.SetContextFromResource(p.Context, check)
	return fetchHandlersWithNames(ctx, r.handlerCtrl, check.OutputMetricHandlers)
}

// ProxyEntityID implements response to request for 'proxyEntityId' field.
func (r *checkCfgImpl) ProxyEntityID(p graphql.ResolveParams) (string, error) {
	check := p.Source.(*types.CheckConfig)
	return check.ProxyEntityID, nil
}

// IsSilenced implements response to request for 'isSilenced' field.
func (r *checkCfgImpl) IsSilenced(p graphql.ResolveParams) (bool, error) {
	check := p.Source.(*types.CheckConfig)
	ctx := types.SetContextFromResource(p.Context, check)
	sls, err := fetchCheckConfigSilences(ctx, r.silenceQuerier, check)
	return len(sls) > 0, err
}

// Silences implements response to request for 'silences' field.
func (r *checkCfgImpl) Silences(p graphql.ResolveParams) (interface{}, error) {
	check := p.Source.(*types.CheckConfig)
	ctx := types.SetContextFromResource(p.Context, check)
	sls, err := fetchCheckConfigSilences(ctx, r.silenceQuerier, check)
	return sls, err
}

// ToJSON implements response to request for 'toJSON' field.
func (r *checkCfgImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	check := p.Source.(*types.CheckConfig)
	return types.WrapResource(check), nil
}

func fetchHandlersWithNames(ctx context.Context, ctrl actions.HandlerController, names []string) ([]*types.Handler, error) {
	handlers, err := ctrl.Query(ctx)
	if err != nil {
		return nil, err
	}

	// Filter out irrevelant handlers
	relevantHandlers := handlers[:0]
	for _, handler := range handlers {
		if !strings.FoundInArray(handler.Name, names) {
			continue
		}
		relevantHandlers = append(relevantHandlers, handler)
	}
	return relevantHandlers, nil
}

func fetchCheckConfigSilences(ctx context.Context, ctrl silenceQuerier, check namedCheck) ([]*types.Silenced, error) {
	sls, err := ctrl.Query(ctx, "", "")
	matched := make([]*types.Silenced, 0, len(sls))
	if err != nil {
		return []*types.Silenced{}, err
	}

	now := time.Now().Unix()
	for _, sl := range sls {
		if !sl.StartSilence(now) {
			continue
		}
		if (sl.Check == check.GetName() && (sl.Subscription == "" || sl.Subscription == "*")) ||
			((sl.Check == "" || sl.Check == "*") && strings.InArray(sl.Subscription, check.GetSubscriptions())) {
			matched = append(matched, sl)
		}
	}

	return matched, nil
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
	handlerCtrl    actions.HandlerController
	silenceQuerier silenceQuerier
}

func newCheckImpl(store store.Store) *checkImpl {
	handlerCtrl := actions.NewHandlerController(store)
	silenceCtrl := actions.NewSilencedController(store)

	return &checkImpl{
		handlerCtrl:    handlerCtrl,
		silenceQuerier: silenceCtrl,
	}
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
		Namespace: check.Namespace,
		Name:      check.Name,
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
	check := p.Source.(*types.Check)
	ctx := types.SetContextFromResource(p.Context, check)
	return fetchHandlersWithNames(ctx, r.handlerCtrl, check.Handlers)
}

// IsSilenced implements response to request for 'isSilenced' field.
func (r *checkImpl) IsSilenced(p graphql.ResolveParams) (bool, error) {
	check := p.Source.(*types.Check)
	return len(check.Silenced) > 0, nil
}

// Silences implements response to request for 'silences' field.
func (r *checkImpl) Silences(p graphql.ResolveParams) (interface{}, error) {
	check := p.Source.(*types.Check)
	ctx := types.SetContextFromResource(p.Context, check)
	sls, err := fetchCheckSilences(ctx, r.silenceQuerier, check)
	return sls, err
}

// OutputMetricHandlers implements response to request for 'outputMetricHandlers' field.
func (r *checkImpl) OutputMetricHandlers(p graphql.ResolveParams) (interface{}, error) {
	check := p.Source.(*types.Check)
	ctx := types.SetContextFromResource(p.Context, check)
	return fetchHandlersWithNames(ctx, r.handlerCtrl, check.OutputMetricHandlers)
}

// ProxyEntityID implements response to request for 'proxyEntityId' field.
func (r *checkImpl) ProxyEntityID(p graphql.ResolveParams) (string, error) {
	check := p.Source.(*types.Check)
	return check.ProxyEntityID, nil
}

func fetchCheckSilences(ctx context.Context, ctrl silenceQuerier, check silenceableCheck) ([]*types.Silenced, error) {
	sls, err := ctrl.Query(ctx, "", "")
	matched := make([]*types.Silenced, 0, len(sls))
	if err != nil {
		return matched, err
	}

	for _, sl := range sls {
		if strings.InArray(sl.ID, check.GetSilenced()) {
			matched = append(matched, sl)
		}
	}

	return matched, nil
}

//
// Implement CheckHistoryFieldResolvers
//

type checkHistoryImpl struct{}

// Status implements response to request for 'status' field.
func (r *checkHistoryImpl) Status(p graphql.ResolveParams) (int, error) {
	h := p.Source.(types.CheckHistory)
	return int(h.Status), nil
}

// Executed implements response to request for 'executed' field.
func (r *checkHistoryImpl) Executed(p graphql.ResolveParams) (time.Time, error) {
	h := p.Source.(types.CheckHistory)
	return time.Unix(h.Executed, 0), nil
}
