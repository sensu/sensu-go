package registry

import (
	"fmt"
	"testing"

	metav1 "github.com/sensu/sensu-go/apis/meta/v1"
	"github.com/sensu/sensu-go/internal/api"
)

func TestRegistryResourceAliases(t *testing.T) {
	for key, kind := range typeRegistry {
		if api.IsInternal(key.APIVersion) {
			continue
		}
		r := kind.(interface{ ResourceName() string })
		t.Run(fmt.Sprintf("%s -> %s", key.Kind, r.ResourceName()), func(t *testing.T) {
			_, ok := typeRegistry[metav1.TypeMeta{APIVersion: key.APIVersion, Kind: r.ResourceName()}]
			if !ok {
				t.Fatalf("%v resource missing", key)
			}
		})
	}
}
