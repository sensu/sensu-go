package handlers

import (
	"context"
	"reflect"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
)

// List ...
func (h Handlers) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	typeOfResource := reflect.TypeOf(h.Resource)
	sliceOfResource := reflect.SliceOf(typeOfResource)
	// Create a pointer and then set the slice value
	ptr := reflect.New(sliceOfResource)
	ptr.Elem().Set(reflect.MakeSlice(sliceOfResource, 0, 0))

	if err := h.Store.ListResource(ctx, h.Resource.StorePath(), ptr.Interface(), pred); err != nil {
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
