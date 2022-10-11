package fixture

import (
	"fmt"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/types"
)

func init() {
	types.RegisterResolver("testing/fixture", func(name string) (interface{}, error) {
		switch name {
		case "V3Resource", "v3_resource":
			return new(V3Resource), nil
		default:
			return nil, fmt.Errorf("invalid resource: %s", name)
		}
	})
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
