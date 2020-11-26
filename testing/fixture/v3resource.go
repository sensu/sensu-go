package fixture

import (
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

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
