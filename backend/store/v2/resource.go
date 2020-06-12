package v2

import (
	"context"

	corev3 "github.com/sensu/sensu-go/api/core/v3"
)

// ResourceRequest contains all the information necessary to query a store.
type ResourceRequest struct {
	Namespace string
	Name      string
	StoreName string
	Context   context.Context
}

// NewResourceRequestFromResource creates a ResourceRequest from a resource.
func NewResourceRequestFromResource(ctx context.Context, r corev3.Resource) ResourceRequest {
	meta := r.GetMetadata()
	if meta == nil {
		return ResourceRequest{}
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
