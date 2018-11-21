package graphql

import (
	"context"
	"sort"
	"time"

	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/strings"
)

var _ schema.EntityFieldResolvers = (*entityImpl)(nil)
var _ schema.SystemFieldResolvers = (*systemImpl)(nil)
var _ schema.NetworkFieldResolvers = (*networkImpl)(nil)
var _ schema.NetworkInterfaceFieldResolvers = (*networkInterfaceImpl)(nil)
var _ schema.DeregistrationFieldResolvers = (*deregistrationImpl)(nil)

//
// Implement EntityFieldResolvers
//

type entityImpl struct {
	schema.EntityAliases
	entityQuerier  entityQuerier
	eventQuerier   eventQuerier
	silenceQuerier silenceQuerier
}

func newEntityImpl(store store.Store) *entityImpl {
	entityCtrl := actions.NewEntityController(store)
	eventCtrl := actions.NewEventController(store, nil)
	silenceCtrl := actions.NewSilencedController(store)

	return &entityImpl{
		entityQuerier:  entityCtrl,
		eventQuerier:   eventCtrl,
		silenceQuerier: silenceCtrl,
	}
}

// ID implements response to request for 'id' field.
func (*entityImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.EntityTranslator.EncodeToString(p.Source), nil
}

// ExtendedAttributes implements response to request for 'extendedAttributes' field.
func (*entityImpl) ExtendedAttributes(p graphql.ResolveParams) (interface{}, error) {
	entity := p.Source.(*types.Entity)
	return wrapExtendedAttributes(entity.ExtendedAttributes), nil
}

// LastSeen implements response to request for 'executed' field.
func (r *entityImpl) LastSeen(p graphql.ResolveParams) (*time.Time, error) {
	e := p.Source.(*types.Entity)
	return convertTs(e.LastSeen), nil
}

// Events implements response to request for 'events' field.
func (r *entityImpl) Events(p schema.EntityEventsFieldResolverParams) (interface{}, error) {
	entity := p.Source.(*types.Entity)

	// fetch
	ctx := types.SetContextFromResource(p.Context, entity)
	evs, err := r.eventQuerier.Query(ctx, entity.Name, "")
	if err != nil {
		return 0, err
	}

	// sort records
	sortEvents(evs, p.Args.OrderBy)

	return evs, nil
}

// Related implements response to request for 'related' field.
func (r *entityImpl) Related(p schema.EntityRelatedFieldResolverParams) (interface{}, error) {
	entity := p.Source.(*types.Entity)

	// fetch
	ctx := types.SetContextFromResource(p.Context, entity)
	entities, err := r.entityQuerier.Query(ctx)
	if err != nil {
		return []*types.Entity{}, err
	}

	// omit source
	for i, en := range entities {
		if en.Name != entity.Name {
			continue
		}

		//
		// - As we sort the result set in the next step we can safely remove the
		//   source from the slice without preserving it's order.
		// - Since we are dealing with a slice of pointers we explicilty set the
		//   reference to the last element to nil to ensure it can be GC'd.
		//
		// https://github.com/golang/go/wiki/SliceTricks#delete-without-preserving-order
		//
		entities[i] = entities[len(entities)-1]
		entities[len(entities)-1] = nil
		entities = entities[:len(entities)-1]
		break
	}

	// sort
	scores := map[int]int{}
	for i, en := range entities {
		matched := strings.Intersect(
			append(en.Subscriptions, en.EntityClass, en.System.Platform),
			append(entity.Subscriptions, entity.EntityClass, entity.System.Platform),
		)
		scores[i] = len(matched)
	}
	sort.Slice(entities, func(i, j int) bool {
		return scores[i] > scores[j]
	})

	// limit
	limit := clampInt(p.Args.Limit, 0, len(entities))
	return entities[0:limit], nil
}

// Status implements response to request for 'status' field.
func (r *entityImpl) Status(p graphql.ResolveParams) (int, error) {
	entity := p.Source.(*types.Entity)

	// fetch
	ctx := types.SetContextFromResource(p.Context, entity)
	evs, err := r.eventQuerier.Query(ctx, entity.Name, "")
	if err != nil {
		return 0, err
	}

	// find MAX value
	var st uint32
	for _, ev := range evs {
		if ev.Check == nil {
			continue
		}
		st = maxUint32(ev.Check.Status, st)
	}
	return int(st), nil
}

// IsSilenced implements response to request for 'isSilenced' field.
func (r *entityImpl) IsSilenced(p graphql.ResolveParams) (bool, error) {
	entity := p.Source.(*types.Entity)
	ctx := types.SetContextFromResource(p.Context, entity)
	sls, err := fetchEntitySilencedEntries(ctx, r.silenceQuerier, entity)
	return len(sls) > 0, err
}

// Silences implements response to request for 'silences' field.
func (r *entityImpl) Silences(p graphql.ResolveParams) (interface{}, error) {
	entity := p.Source.(*types.Entity)
	ctx := types.SetContextFromResource(p.Context, entity)
	sls, err := fetchEntitySilencedEntries(ctx, r.silenceQuerier, entity)
	return sls, err
}

func fetchEntitySilencedEntries(ctx context.Context, ctrl silenceQuerier, entity *types.Entity) ([]*types.Silenced, error) {
	sls, err := ctrl.Query(ctx, "", "")
	matched := make([]*types.Silenced, 0, len(sls))
	if err != nil {
		return matched, err
	}

	now := time.Now().Unix()
	for _, sl := range sls {
		if !(sl.Check == "" || sl.Check == "*") || !sl.StartSilence(now) {
			continue
		}
		if strings.InArray(sl.Subscription, entity.Subscriptions) {
			matched = append(matched, sl)
		}
	}

	return matched, nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*entityImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.Entity)
	return ok
}

//
// Implement SystemFieldResolvers
//

type systemImpl struct {
	schema.SystemAliases
}

// Os implements response to request for 'os' field.
func (r *systemImpl) Os(p graphql.ResolveParams) (string, error) {
	sys := p.Source.(types.System)
	return sys.OS, nil
}

//
// Implement NetworkFieldResolvers
//

type networkImpl struct {
	schema.NetworkAliases
}

//
// Implement NetworkInterfaceFieldResolvers
//

type networkInterfaceImpl struct {
	schema.NetworkInterfaceAliases
}

// Mac implements response to request for 'mac' field.
func (r *networkInterfaceImpl) Mac(p graphql.ResolveParams) (string, error) {
	i := p.Source.(types.NetworkInterface)
	return i.MAC, nil
}

//
// Implement DeregistrationFieldResolvers
//

type deregistrationImpl struct {
	schema.DeregistrationAliases
}
