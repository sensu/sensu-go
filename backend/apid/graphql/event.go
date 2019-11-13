package graphql

import (
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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
	return globalid.EventTranslator.EncodeToString(p.Context, p.Source), nil
}

// Namespace implements response to request for 'namespace' field.
func (r *eventImpl) Namespace(p graphql.ResolveParams) (string, error) {
	event := p.Source.(*corev2.Event)
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
	event := p.Source.(*corev2.Event)
	return time.Unix(event.Timestamp, 0), nil
}

// IsIncident implements response to request for 'isIncident' field.
func (r *eventImpl) IsIncident(p graphql.ResolveParams) (bool, error) {
	event := p.Source.(*corev2.Event)
	return event.IsIncident(), nil
}

// IsNewIncident implements response to request for 'isNewIncident' field.
func (r *eventImpl) IsNewIncident(p graphql.ResolveParams) (bool, error) {
	event := p.Source.(*corev2.Event)
	check := event.Check
	if check == nil || check.Status == 0 || len(check.History) == 0 {
		return false, nil
	}

	lastExecution := check.History[0].Status
	return lastExecution == 0, nil
}

// IsResolution implements response to request for 'isResolution' field.
func (r *eventImpl) IsResolution(p graphql.ResolveParams) (bool, error) {
	event := p.Source.(*corev2.Event)
	return event.IsResolution(), nil
}

// WasSilenced implements response to request for 'wasSilenced' field.
func (r *eventImpl) WasSilenced(p graphql.ResolveParams) (bool, error) {
	event := p.Source.(*corev2.Event)
	return event.IsSilenced(), nil
}

// IsSilenced implements response to request for 'isSilenced' field.
func (r *eventImpl) IsSilenced(p graphql.ResolveParams) (bool, error) {
	src := p.Source.(*corev2.Event)
	results, err := loadSilenceds(p.Context, src.Namespace)
	records := filterSilenceds(results, src.IsSilencedBy)
	return len(records) > 0, err
}

// Silences implements response to request for 'silences' field.
func (r *eventImpl) Silences(p graphql.ResolveParams) (interface{}, error) {
	src := p.Source.(*corev2.Event)
	results, err := loadSilenceds(p.Context, src.Namespace)
	records := filterSilenceds(results, src.IsSilencedBy)

	return records, err
}

// IsTypeOf is used to determine if a given value is associated with the type
func (r *eventImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.Event)
	return ok
}

// ToJSON implements response to request for 'toJSON' field.
func (r *eventImpl) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	return types.WrapResource(p.Source.(corev2.Resource)), nil
}
