package graphql

import (
	"time"

	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.EventFieldResolvers = (*eventImpl)(nil)

//
// Implement CheckConfigFieldResolvers
//

type eventImpl struct {
	schema.EventAliases
}

// ID implements response to request for 'id' field.
func (r *eventImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.EventTranslator.EncodeToString(p.Source), nil
}

// Namespace implements response to request for 'namespace' field.
func (r *eventImpl) Namespace(p graphql.ResolveParams) (string, error) {
	event := p.Source.(*types.Event)
	if event.Entity != nil {
		return event.Entity.Namespace, nil
	}
	if event.Check != nil {
		return event.Check.Namespace, nil
	}
	return "", nil
}

// Timestamp implements response to request for 'timestamp' field.
func (r *eventImpl) Timestamp(p graphql.ResolveParams) (time.Time, error) {
	event := p.Source.(*types.Event)
	return time.Unix(event.Timestamp, 0), nil
}

// IsIncident implements response to request for 'isIncident' field.
func (r *eventImpl) IsIncident(p graphql.ResolveParams) (bool, error) {
	event := p.Source.(*types.Event)
	return event.IsIncident(), nil
}

// IsNewIncident implements response to request for 'isNewIncident' field.
func (r *eventImpl) IsNewIncident(p graphql.ResolveParams) (bool, error) {
	event := p.Source.(*types.Event)
	check := event.Check
	if check == nil || check.Status == 0 || len(check.History) == 0 {
		return false, nil
	}

	lastExecution := check.History[0].Status
	return lastExecution == 0, nil
}

// IsResolution implements response to request for 'isResolution' field.
func (r *eventImpl) IsResolution(p graphql.ResolveParams) (bool, error) {
	event := p.Source.(*types.Event)
	return event.IsResolution(), nil
}

// IsSilenced implements response to request for 'isSilenced' field.
func (r *eventImpl) IsSilenced(p graphql.ResolveParams) (bool, error) {
	event := p.Source.(*types.Event)
	return event.IsSilenced(), nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (r *eventImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*types.Event)
	return ok
}
