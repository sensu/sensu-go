package types

import (
	"sync"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

var fieldsetFnMapMu = &sync.RWMutex{}
var fieldsetFnMap = map[string]FieldsetFnGetter{
	"core/v2": corev2.LookupFieldsetFn,
}

// RegisterFieldsetGetter ...
type FieldsetFnGetter func(typename string) (corev2.FieldsetFn, bool)

// RegisterFieldsetFn registers a fieldset getter.
func RegisterFieldsetFn(apiVersion string, getter FieldsetFnGetter) {
	fieldsetFnMapMu.Lock()
	defer fieldsetFnMapMu.Unlock()
	fieldsetFnMap[apiVersion] = getter
}

// LookupFieldsetFn can be used to lookup a fieldset given a api version and
// typename.
func LookupFieldsetFn(apiVersion string, typename string) (corev2.FieldsetFn, bool) {
	fieldsetFnMapMu.Lock()
	defer fieldsetFnMapMu.Unlock()
	getter, ok := fieldsetFnMap[apiVersion]
	if !ok {
		return nil, false
	}
	return getter(typename)
}
