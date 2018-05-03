package graphql

import (
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
	userCtrl   actions.UserController
	entityCtrl entityQuerier
	eventCtrl  eventQuerier
}

func newEntityImpl(store store.Store) *entityImpl {
	userCtrl := actions.NewUserController(store)
	entityCtrl := actions.NewEntityController(store)
	eventCtrl := actions.NewEventController(store, nil)

	return &entityImpl{
		userCtrl:   userCtrl,
		entityCtrl: entityCtrl,
		eventCtrl:  eventCtrl,
	}
}

// ID implements response to request for 'id' field.
func (*entityImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.EntityTranslator.EncodeToString(p.Source), nil
}

// Namespace implements response to request for 'namespace' field.
func (*entityImpl) Namespace(p graphql.ResolveParams) (interface{}, error) {
	return p.Source, nil
}

// Name implements response to request for 'name' field.
func (*entityImpl) Name(p graphql.ResolveParams) (string, error) {
	entity := p.Source.(*types.Entity)
	return entity.ID, nil
}

// AuthorId implements response to request for 'authorId' field.
func (*entityImpl) AuthorID(p graphql.ResolveParams) (string, error) {
	entity := p.Source.(*types.Entity)
	return entity.User, nil
}

// LastSeen implements response to request for 'executed' field.
func (r *entityImpl) LastSeen(p graphql.ResolveParams) (*time.Time, error) {
	c := p.Source.(*types.Entity)
	if c.LastSeen == 0 {
		return nil, nil
	}
	t := time.Unix(c.LastSeen, 0)
	return &t, nil
}

// Author implements response to request for 'author' field.
func (r *entityImpl) Author(p graphql.ResolveParams) (interface{}, error) {
	entity := p.Source.(*types.Entity)
	user, err := r.userCtrl.Find(p.Context, entity.User)
	return handleControllerResults(user, err)
}

// Related implements response to request for 'related' field.
func (r *entityImpl) Related(p schema.EntityRelatedFieldResolverParams) (interface{}, error) {
	entity := p.Source.(*types.Entity)

	// fetch
	ctx := types.SetContextFromResource(p.Context, entity)
	entities, err := r.entityCtrl.Query(ctx)
	if err != nil {
		return []*types.Entity{}, err
	}

	// omit source
	for i, en := range entities {
		if en.ID != entity.ID {
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
			append(en.Subscriptions, en.Class, en.System.Platform),
			append(entity.Subscriptions, entity.Class, entity.System.Platform),
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
	evs, err := r.eventCtrl.Query(ctx, entity.ID, "")
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
