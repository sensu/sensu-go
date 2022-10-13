package fixture

import (
	"fmt"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-api-tools/apis"
)

func init() {
	apis.RegisterType("testing/fixture", new(V3Resource), apis.WithAlias("v3_resource"))
}

type V3Resource struct {
	Metadata *corev2.ObjectMeta
}

func (v *V3Resource) GetMetadata() *corev2.ObjectMeta {
	return v.Metadata
}

func (v *V3Resource) SetMetadata(m *corev2.ObjectMeta) {
	v.Metadata = m
}

func (v *V3Resource) StoreName() string {
	return "v3resource"
}

func (v *V3Resource) RBACName() string {
	return "v3resource"
}

func (v *V3Resource) URIPath() string {
	return fmt.Sprintf("/v3resource/%s", v.Metadata.Name)
}

func (v *V3Resource) Validate() error {
	return nil
}
