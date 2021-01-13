package v2

import (
	"context"
	"errors"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
)

type SortOrder int

const (
	SortNone SortOrder = iota
	SortAscend
	SortDescend
)

// ResourceRequest contains all the information necessary to query a store.
type ResourceRequest struct {
	Namespace string
	Name      string
	StoreName string
	Context   context.Context
	SortOrder SortOrder
}

// NewResourceRequestFromResource creates a ResourceRequest from a resource.
func NewResourceRequestFromResource(ctx context.Context, r corev3.Resource) ResourceRequest {
	meta := r.GetMetadata()
	if meta == nil {
		meta = &corev2.ObjectMeta{}
	}
	return ResourceRequest{
		Namespace: meta.Namespace,
		Name:      meta.Name,
		StoreName: r.StoreName(),
		Context:   ctx,
	}
}

// NewResourceRequest creates a resource request.
func NewResourceRequest(ctx context.Context, namespace, name, storeName string) ResourceRequest {
	return ResourceRequest{
		Namespace: namespace,
		Name:      name,
		StoreName: storeName,
		Context:   ctx,
	}
}

func NewResourceRequestFromV2Resource(ctx context.Context, r corev2.Resource) ResourceRequest {
	meta := r.GetObjectMeta()
	return ResourceRequest{
		Namespace: meta.Namespace,
		Name:      meta.Name,
		StoreName: r.StorePrefix(),
		Context:   ctx,
	}
}

func (r ResourceRequest) Validate() error {
	if r.StoreName == "" {
		return errors.New("invalid resource request: missing store name")
	}
	if r.Context == nil {
		return errors.New("invalid resource request: nil context")
	}
	return nil
}
