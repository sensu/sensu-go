package handlers

import (
	"context"
	"reflect"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/core/v3/types"
	"github.com/sensu/sensu-go/backend/apid/request"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

func (h Handlers[R, T]) ListResources(ctx context.Context, pred *store.SelectionPredicate) ([]corev3.Resource, error) {
	namespace := corev2.ContextNamespace(ctx)
	gstore := storev2.Of[R](h.Store)

	if selector := request.SelectorFromContext(ctx); selector != nil {
		tmp := new(T)
		var tm corev2.TypeMeta
		if getter, ok := any(tmp).(tmGetter); ok {
			tm = getter.GetTypeMeta()
		} else {
			typ := reflect.Indirect(reflect.ValueOf(tmp)).Type()
			tm = corev2.TypeMeta{
				Type:       typ.Name(),
				APIVersion: types.ApiVersion(typ.PkgPath()),
			}
		}
		ctx = storev2.ContextWithSelector(ctx, tm, selector)
	}

	list, err := gstore.List(ctx, storev2.ID{Namespace: namespace}, pred)
	if err != nil {
		return nil, err
	}

	result := make([]corev3.Resource, len(list))
	for i := range list {
		var r R = list[i]
		result[i] = r
	}
	return result, nil
}

type tmGetter interface {
	GetTypeMeta() corev2.TypeMeta
}
