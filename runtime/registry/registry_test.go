package registry

import (
	"fmt"
	"reflect"
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

func TestResolveSlice(t *testing.T) {
	for key, kind := range typeRegistry {
		t.Run(fmt.Sprintf("slice of %s", key.Kind), func(t *testing.T) {
			defer func() {
				if e := recover(); e != nil {
					t.Fatal(e)
				}
			}()
			slice, err := ResolveSlice(key)
			if err != nil {
				t.Fatal(err)
			}
			// Will panic if ResolveSlice is broken
			reflect.Append(reflect.ValueOf(slice), reflect.ValueOf(kind))
		})
	}
}
