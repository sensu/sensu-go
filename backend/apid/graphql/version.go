package graphql

import (
	"time"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/version"
)

var _ schema.VersionsFieldResolvers = (*versionsImpl)(nil)
var _ schema.SensuBackendVersionFieldResolvers = (*sensuBackendVersionImpl)(nil)

//
// Implement VersionsFieldResolvers
//

type versionsImpl struct {
	schema.VersionsAliases
}

// Backend implements response to request for 'backend' field.
func (r *versionsImpl) Backend(p graphql.ResolveParams) (interface{}, error) {
	return struct{}{}, nil
}

//
// Implement SensuBackendVersionFieldResolvers
//

type sensuBackendVersionImpl struct{}

// Version implements response to request for 'version' field.
func (r *sensuBackendVersionImpl) Version(p graphql.ResolveParams) (string, error) {
	return version.Semver(), nil
}

// BuildSha implements response to request for 'version' field.
func (r *sensuBackendVersionImpl) BuildSHA(p graphql.ResolveParams) (string, error) {
	return version.BuildSHA, nil
}

// BuildDate implements response to request for 'buildDate' field.
func (r *sensuBackendVersionImpl) BuildDate(p graphql.ResolveParams) (*time.Time, error) {
	if t, err := time.Parse(time.RFC3339, version.BuildDate); err == nil {
		return &t, nil
	}
	return nil, nil
}
