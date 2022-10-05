package v3

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

var entityStateRBACName = (&corev2.Entity{}).RBACName()

func (*EntityState) rbacName() string {
	return entityStateRBACName
}
