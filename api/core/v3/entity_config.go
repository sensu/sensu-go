package v3

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

var entityConfigRBACName = (&corev2.Entity{}).RBACName()

func (*EntityConfig) rbacName() string {
	return entityConfigRBACName
}
