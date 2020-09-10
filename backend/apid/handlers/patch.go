package handlers

import (
	"encoding/json"
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

// PatchResource patches a given resource, using the request body as the patch
func (h Handlers) PatchResource(r *http.Request) (interface{}, error) {
	// First read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, actions.NewError(
			actions.InvalidArgument,
			fmt.Errorf("could not read the request body: %s", err),
		)
	}
	// TODO(palourde): Retrieve the etag header here

	// Initialize our patcher with the body
	patcher := &patch.Merge{JSONPatch: body}

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

	objectMeta := corev2.ObjectMeta{}

	// Add the namespace & namespace in the patch if they are not defined, so we
	// pass the metadata validation
	if resource.GetObjectMeta().Name == "" {
		objectMeta.Name = name
	}
	if resource.GetObjectMeta().Namespace == "" {
		objectMeta.Namespace = namespace
	}
	resource.SetObjectMeta(objectMeta)

	if err := CheckMeta(payload.Interface(), mux.Vars(r), "id"); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	// resource, ok := payload.Interface().(corev2.Resource)
	// if !ok {
	// 	fmt.Println("3")
	// 	return nil, actions.NewErrorf(actions.InvalidArgument)
	// }

	// Retrieve the name of the resource via the route variables
	// params := mux.Vars(r)
	// name, err := url.PathUnescape(params["id"])
	// if err != nil {
	// 	fmt.Println("4")
	// 	return nil, err
	// }

	// TODO(palourde): Deal with the new etag here
	_, err = h.Store.PatchResource(r.Context(), resource, name, patcher, []byte{})
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

	// TODO(palourde): We could compare the etag and return a 200 OK when we had a
	// modification and a 204 No Content when they are the same
	return resource, nil
}
