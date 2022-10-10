package handlers

import (
	"context"
	"reflect"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
)

// ListResources lists all resources for the resource type
func (h Handlers) ListResources(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	// Get the type of the resource and create a slice type of []type
	typeOfResource := reflect.TypeOf(h.Resource)
	sliceOfResource := reflect.SliceOf(typeOfResource)
	// Create a pointer to our slice type and then set the slice value
	ptr := reflect.New(sliceOfResource)
	ptr.Elem().Set(reflect.MakeSlice(sliceOfResource, 0, 0))

	if err := h.Store.ListResources(ctx, h.Resource.StorePrefix(), ptr.Interface(), pred); err != nil {
		return nil, actions.NewError(actions.InternalErr, err)
	}

	results := ptr.Elem()
	resources := make([]corev2.Resource, results.Len())
	for i := 0; i < results.Len(); i++ {
		r, ok := results.Index(i).Interface().(corev2.Resource)
		if !ok {
			logger.Errorf("%T is not core2.Resource", results.Index(i).Interface())
			continue
		}
		resources[i] = r
	}

	return resources, nil
}
