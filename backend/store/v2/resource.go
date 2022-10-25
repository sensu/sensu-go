package v2

import (
	"errors"
	"path"
	"reflect"
	"strings"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
)

type SortOrder int

const (
	SortNone SortOrder = iota
	SortAscend
	SortDescend
)

// tmGetter is useful for types that want to explicitly provide their TypeMeta
type tmGetter interface {
	GetTypeMeta() corev2.TypeMeta
}

func getTypeMeta(r interface{}) corev2.TypeMeta {
	var tm corev2.TypeMeta
	if getter, ok := r.(tmGetter); ok {
		tm = getter.GetTypeMeta()
	} else {
		typ := reflect.Indirect(reflect.ValueOf(r)).Type()
		tm = corev2.TypeMeta{
			Type:       typ.Name(),
			APIVersion: apiVersion(typ.PkgPath()),
		}
	}
	return tm
}

func apiVersion(version string) string {
	parts := strings.Split(version, "/")
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	return path.Join(parts[len(parts)-2], parts[len(parts)-1])
}

// WrapResource is made variable, for the purpose of swapping it out for another
// implementation.
var WrapResource = func(resource corev3.Resource, opts ...wrap.Option) (Wrapper, error) {
	return wrap.Resource(resource, opts...)
}

// ResourceRequest contains all the information necessary to query a store.
type ResourceRequest struct {
	Type       string
	APIVersion string
	Namespace  string
	Name       string
	StoreName  string
	SortOrder  SortOrder
}

// NewResourceRequestFromResource creates a ResourceRequest from a resource.
func NewResourceRequestFromResource(r corev3.Resource) ResourceRequest {
	meta := r.GetMetadata()
	if meta == nil {
		meta = &corev2.ObjectMeta{}
	}
	tm := getTypeMeta(r)
	return ResourceRequest{
		Type:       tm.Type,
		APIVersion: tm.APIVersion,
		Namespace:  meta.Namespace,
		Name:       meta.Name,
		StoreName:  r.StoreName(),
	}
}

// NewResourceRequest creates a resource request.
func NewResourceRequest(meta corev2.TypeMeta, namespace, name, storeName string) ResourceRequest {
	return ResourceRequest{
		Type:       meta.Type,
		APIVersion: meta.APIVersion,
		Namespace:  namespace,
		Name:       name,
		StoreName:  storeName,
	}
}

func NewResourceRequestFromV2Resource(r corev2.Resource) ResourceRequest {
	meta := r.GetObjectMeta()
	tm := getTypeMeta(r)
	return ResourceRequest{
		Type:       tm.Type,
		APIVersion: tm.APIVersion,
		Namespace:  meta.Namespace,
		Name:       meta.Name,
		StoreName:  r.StorePrefix(),
	}
}

func (r ResourceRequest) Validate() error {
	if r.StoreName == "" {
		return errors.New("invalid resource request: missing store name")
	}
	if r.APIVersion == "" || r.Type == "" {
		return errors.New("invalid resource request: empty type metadata")
	}
	return nil
}
