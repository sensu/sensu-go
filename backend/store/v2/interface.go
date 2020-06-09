package v2

import (
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/wrap"
)

type ResourceRequest struct {
	Namespace   string
	Name        string
	StoreSuffix string
}

func NewResourceRequestFromResource(r corev3.Resource) ResourceRequest {
	meta := r.GetMetadata()
	if meta == nil {
		return ResourceRequest{}
	}
	return ResourceRequest{
		Namespace:   meta.Namespace,
		Name:        meta.Name,
		StoreSuffix: r.StoreSuffix(),
	}
}

func NewResourceRequest(namespace, name, storeSuffix string) ResourceRequest {
	return ResourceRequest{
		Namespace:   namespace,
		Name:        name,
		StoreSuffix: storeSuffix,
	}
}

type Interface interface {
	CreateOrUpdate(ResourceRequest, *wrap.Wrapper) error
	UpdateIfNotExists(ResourceRequest, *wrap.Wrapper) error
	CreateIfNotExists(ResourceRequest, *wrap.Wrapper) error
	Get(ResourceRequest) (*wrap.Wrapper, error)
	GetIfExists(ResourceRequest) (*wrap.Wrapper, error)
	Delete(ResourceRequest) error
	DeleteIfExists(ResourceRequest) error
	List(ResourceRequest, *store.SelectionPredicate) ([]*wrap.Wrapper, error)
}
