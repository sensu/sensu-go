package graphql

import (
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
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
func (*entityImpl) ID(p graphql.ResolveParams) (interface{}, error) {
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

// IsTypeOf is used to determine if a given value is associated with the type
func (*systemImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(types.System)
	return ok
}

//
// Implement NetworkFieldResolvers
//

type networkImpl struct {
	schema.NetworkAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*networkImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(types.Network)
	return ok
}

//
// Implement NetworkInterfaceFieldResolvers
//

type networkInterfaceImpl struct {
	schema.NetworkInterfaceAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*networkInterfaceImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(types.NetworkInterface)
	return ok
}

//
// Implement DeregistrationFieldResolvers
//

type deregistrationImpl struct {
	schema.DeregistrationAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*deregistrationImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	_, ok := s.(types.Deregistration)
	return ok
}
