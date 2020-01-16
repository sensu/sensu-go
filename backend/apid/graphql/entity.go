package graphql

import (
	"sort"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
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
}

// ID implements response to request for 'id' field.
func (*entityImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.EntityTranslator.EncodeToString(p.Context, p.Source), nil
}

// Metadata implements response to request for 'metadata' field.
func (*entityImpl) Metadata(p graphql.ResolveParams) (interface{}, error) {
	entity := p.Source.(*corev2.Entity)
	entity = entity.GetRedactedEntity()
	return entity.GetObjectMeta(), nil
}

// LastSeen implements response to request for 'executed' field.
func (r *entityImpl) LastSeen(p graphql.ResolveParams) (*time.Time, error) {
	e := p.Source.(*corev2.Entity)
	return convertTs(e.LastSeen), nil
}

// Events implements response to request for 'events' field.
func (r *entityImpl) Events(p schema.EntityEventsFieldResolverParams) (interface{}, error) {
	src := p.Source.(*corev2.Entity)

	// fetch
	results, err := loadEvents(p.Context, src.Namespace)
	if err != nil {
		return []interface{}{}, err
	}

	// filter
	matches, err := filter.Compile(p.Args.Filters, EventFilters(), corev2.EventFields)
	if err != nil {
		return []interface{}{}, err
	}
	records := filterEvents(results, func(obj *corev2.Event) bool {
		return obj.Entity.Name == src.Name && matches(obj)
	})

	// sort records
	sortEvents(records, p.Args.OrderBy)

	return records, nil
}

// Related implements response to request for 'related' field.
func (r *entityImpl) Related(p schema.EntityRelatedFieldResolverParams) (interface{}, error) {
	// fetch
	entity := p.Source.(*corev2.Entity)
	results, err := loadEntities(p.Context, entity.Namespace)
	if err != nil {
		return []interface{}{}, err
	}

	// omit self
	records := filterEntities(results, func(obj *corev2.Entity) bool {
		return obj.Name != entity.Name
	})

	// sort
	scores := map[int]int{}
	for i, en := range records {
		matched := strings.Intersect(
			append(en.Subscriptions, en.EntityClass, en.System.Platform),
			append(entity.Subscriptions, entity.EntityClass, entity.System.Platform),
		)
		scores[i] = len(matched)
	}
	sort.Slice(records, func(i, j int) bool {
		return scores[i] > scores[j]
	})

	// limit
	limit := clampInt(p.Args.Limit, 0, len(records))
	return records[0:limit], nil
}

// Status implements response to request for 'status' field.
func (r *entityImpl) Status(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*corev2.Entity)

	// fetch
	results, err := loadEvents(p.Context, src.Namespace)
	if err != nil {
		return 0, err
	}

	// filter events associated w/ entity
	evs := filterEvents(results, func(obj *corev2.Event) bool {
		return obj.Entity.Name == src.Name
	})

	// find MAX value
	var st uint32
	for _, ev := range evs {
		if ev.Check == nil {
			continue
		}
		st = maxUint32(ev.Check.Status, st)
	}
	return st, nil
}

// IsSilenced implements response to request for 'isSilenced' field.
func (r *entityImpl) IsSilenced(p graphql.ResolveParams) (bool, error) {
	src := p.Source.(*corev2.Entity)
	results, err := loadSilenceds(p.Context, src.Namespace)
	records := filterSilenceds(results, filterSilenceByEntity(src))
	return len(records) > 0, err
}

// Silences implements response to request for 'silences' field.
func (r *entityImpl) Silences(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*corev2.Entity)
	results, err := loadSilenceds(p.Context, src.Namespace)
	records := filterSilenceds(results, filterSilenceByEntity(src))
	return records, err
}

func filterSilenceByEntity(src *corev2.Entity) silencePredicate {
	now := time.Now().Unix()
	return func(obj *corev2.Silenced) bool {
		if !(obj.Check == "" || obj.Check == "*") || !obj.StartSilence(now) {
			return false
		}
		if strings.InArray(obj.Subscription, src.Subscriptions) {
			return true
		}
		return false
	}
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*entityImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.Entity)
	return ok
}

// ToJSON implements response to request for 'toJSON' field.
func (*entityImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(corev2.Resource)), nil
}

//
// Implement SystemFieldResolvers
//

type systemImpl struct {
	schema.SystemAliases
}

// Os implements response to request for 'os' field.
func (r *systemImpl) Os(p graphql.ResolveParams) (string, error) {
	sys := p.Source.(corev2.System)
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
	i := p.Source.(corev2.NetworkInterface)
	return i.MAC, nil
}

//
// Implement DeregistrationFieldResolvers
//

type deregistrationImpl struct {
	schema.DeregistrationAliases
}
