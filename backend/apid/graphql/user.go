package graphql

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
)

var _ schema.UserFieldResolvers = (*userImpl)(nil)

//
// Implement UserFieldResolvers
//

type userImpl struct {
	schema.UserAliases
}

// ID implements response to request for 'id' field.
func (*userImpl) ID(p graphql.ResolveParams) (string, error) {
	return globalid.UserTranslator.EncodeToString(p.Context, p.Source), nil
}

// AuthorId implements response to request for 'hasPassword' field.
func (*userImpl) HasPassword(p graphql.ResolveParams) (bool, error) {
	user := p.Source.(*corev2.User)
	return len(user.Password) > 0, nil
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*userImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(*corev2.User)
	return ok
}
