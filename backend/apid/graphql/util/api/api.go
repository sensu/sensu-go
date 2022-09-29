package util_api

import (
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// UnwrapListResult from API client, helpful when resolving a field as GraphQL
// does not consider the absence of a value an error; as such we omit the error
// if the API client returns Permission denied.
func UnwrapListResult(res interface{}, err error) (interface{}, error) {
	// legacy resolvers returned empty results when encountering request errs
	if fe := ToFetchErr(err); fe != nil {
		logger.WithField("type", fmt.Sprintf("%T", res)).WithError(err).Info("couldn't access resource")
		return []interface{}{}, nil
	}
	if err != nil {
		return []interface{}{}, err
	}
	return UnwrapList(res), err
}

// UnwrapGetResult from API client, helpful when resolving a field as GraphQL
// does not consider the absence of a value an error; as such we omit the
// error when the API client returns NotFound or Permission denied.
func UnwrapGetResult(res interface{}, err error) (interface{}, error) {
	// legacy resolvers returned empty results when encountering request errs
	if fe := ToFetchErr(err); fe != nil {
		logger.WithField("type", fmt.Sprintf("%T", res)).WithError(err).Info("couldn't access resource")
		return nil, nil
	}
	if _, ok := err.(*store.ErrNotFound); ok {
		logger.WithField("type", fmt.Sprintf("%T", res)).WithError(err).Info("couldn't find resource")
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return UnwrapResource(res), err
}

// HandleListResult helps format the result of a list operation for selection.
func HandleListResult(res interface{}, err error) (interface{}, error) {
	if fe := ToFetchErr(err); fe != nil {
		return fe, nil
	}
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"nodes": UnwrapList(res),
	}, nil
}

// HandleGetResult helps format the result of a get operation for selection.
func HandleGetResult(res interface{}, err error) (interface{}, error) {
	if fe := ToFetchErr(err); fe != nil {
		return fe, nil
	}
	if _, ok := err.(*store.ErrNotFound); ok {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return UnwrapResource(res), err
}

// Unwraps the results returned by the api client.
func UnwrapList(ws interface{}) interface{} {
	if wrapped, ok := ws.([]*types.Wrapper); ok {
		rs := make([]interface{}, len(wrapped))
		for i, w := range wrapped {
			rs[i] = w.Value
		}
		return rs
	}
	if wrapped, ok := ws.([]*corev3.V2ResourceProxy); ok {
		rs := make([]interface{}, len(wrapped))
		for i, w := range wrapped {
			rs[i] = w.Resource
		}
		return rs
	}
	if ws == nil {
		return []interface{}{}
	}
	return ws
}

// UnwrapResource unwrap the proxy or wrapper type from a given resource.
func UnwrapResource(r interface{}) interface{} {
	switch r := r.(type) {
	case *corev3.V2ResourceProxy:
		return r.Resource
	case *types.Wrapper:
		return r.Value
	}
	return r
}

// WrapResource safely wraps the given resource in a type wrapper
func WrapResource(r interface{}) types.Wrapper {
	switch r := r.(type) {
	case corev2.Resource:
		return types.WrapResource(r)
	case corev3.Resource: // maybe we move this into the compat package
		var tm types.TypeMeta
		if getter, ok := r.(interface{ GetTypeMeta() types.TypeMeta }); ok {
			tm = getter.GetTypeMeta()
		}
		var meta corev2.ObjectMeta
		if r.GetMetadata() != nil {
			meta = *r.GetMetadata()
		}
		return types.Wrapper{
			TypeMeta:   tm,
			ObjectMeta: meta,
			Value:      r,
		}
	case *types.Wrapper:
		if r == nil {
			return types.Wrapper{}
		}
		return *r
	case types.Wrapper:
		return r
	}
	panic("wrap error: unknown type")
}

// ToFetchErr produces a FetchErr from the given err; may return nil if the err
// does not have a relavant err code.
func ToFetchErr(err error) map[string]interface{} {
	if err == authorization.ErrUnauthorized || err == authorization.ErrNoClaims {
		return map[string]interface{}{
			"code":    "ERR_PERMISSION_DENIED",
			"message": err.Error(),
		}
	}
	return nil
}
