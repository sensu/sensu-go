package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
)

const (
	mergePatchContentType = "application/merge-patch+json"
	jsonPatchContentType  = "application/json-patch+json"

	ifMatchHeader     = "If-Match"
	ifNoneMatchHeader = "If-None-Match"
)

// PatchResource patches a given resource, using the request body as the patch
func (h Handlers) PatchResource(r *http.Request) (interface{}, error) {
	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, actions.NewError(
			actions.InvalidArgument,
			fmt.Errorf("could not read the request body: %s", err),
		)
	}

	var patcher patch.Patcher

	// Determine the requested PATCH operation based on the Content-Type header
	// and initialize a patcher
	switch contentType := r.Header.Get("Content-Type"); contentType {
	case mergePatchContentType, "": // Use merge patch as fallback value
		patcher = &patch.Merge{MergePatch: body}
	case jsonPatchContentType:
		return nil, actions.NewError(actions.InvalidArgument, errors.New("JSON Patch is not supported yet"))
	default:
		return nil, actions.NewError(actions.InvalidArgument, fmt.Errorf("invalid Content-Type header: %q", contentType))
	}

	// We also need to decode the request body into a concrete type so we can
	// guard against namespace & name alterations
	payload := reflect.New(reflect.TypeOf(h.Resource).Elem())
	if err := json.Unmarshal(body, payload.Interface()); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	resource, ok := payload.Interface().(corev2.Resource)
	if !ok {
		return nil, actions.NewErrorf(actions.InvalidArgument)
	}

	// Retrieve the name & namespace of the resource via the route variables
	params := mux.Vars(r)
	name, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	namespace, err := url.PathUnescape(params["namespace"])
	if err != nil {
		return nil, err
	}

	// Add the namespace & namespace in the patch if they are not defined, so we
	// pass the metadata validation
	objectMeta := resource.GetObjectMeta()
	if objectMeta.Name == "" {
		objectMeta.Name = name
	}
	if objectMeta.Namespace == "" {
		objectMeta.Namespace = namespace
	}
	resource.SetObjectMeta(objectMeta)

	// Validate the metadata
	if err := CheckMeta(resource, mux.Vars(r), "id"); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	// Determine if we have a conditional request
	conditions := &store.ETagCondition{
		IfMatch:     r.Header.Get(ifMatchHeader),
		IfNoneMatch: r.Header.Get(ifNoneMatchHeader),
	}

	err = h.Store.PatchResource(r.Context(), resource, name, patcher, conditions)
	if err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, actions.NewError(actions.NotFound, err)
		case *store.ErrNotValid:
			return nil, actions.NewError(actions.InvalidArgument, err)
		case *store.ErrModified:
			return nil, actions.NewError(actions.PreconditionFailed, err)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}

	return resource, nil
}
