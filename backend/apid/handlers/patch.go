package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
)

const (
	mergePatchContentType = "application/merge-patch+json"
	jsonPatchContentType  = "application/json-patch+json"

	ifMatchHeader     = "If-Match"
	ifNoneMatchHeader = "If-None-Match"
)

// acceptedContentTypes contains the list of content types we accept
var acceptedContentTypes = []string{mergePatchContentType}

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
		return nil, actions.NewError(
			actions.InvalidArgument,
			fmt.Errorf("JSON Patch is not supported yet. Allowed values: %s", strings.Join(acceptedContentTypes, ", ")),
		)
	default:
		return nil, actions.NewError(
			actions.InvalidArgument,
			fmt.Errorf("invalid Content-Type header: %s.  Allowed values: %s", contentType, strings.Join(acceptedContentTypes, ", ")),
		)
	}

	// Determine if we have a conditional request
	conditions := &store.ETagCondition{
		IfMatch:     r.Header.Get(ifMatchHeader),
		IfNoneMatch: r.Header.Get(ifNoneMatchHeader),
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

	// Validate that the patch does not alter the namespace nor the name
	if err := validatePatch(body, params); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	if h.Resource != nil {
		return h.patchV2Resource(r.Context(), body, name, patcher, conditions)
	} else if h.V3Resource != nil {
		return h.patchV3Resource(r.Context(), body, name, namespace, patcher, conditions)
	}

	return nil, actions.NewError(actions.InvalidArgument, errors.New("no resource available"))
}

func (h Handlers) patchV2Resource(ctx context.Context, body []byte, name string, patcher patch.Patcher, conditions *store.ETagCondition) (interface{}, error) {
	payload := reflect.New(reflect.TypeOf(h.Resource).Elem())
	if err := json.Unmarshal(body, payload.Interface()); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	resource, ok := payload.Interface().(corev2.Resource)
	if !ok {
		return nil, actions.NewErrorf(actions.InvalidArgument)
	}

	if err := h.Store.PatchResource(ctx, resource, name, patcher, conditions); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, actions.NewError(actions.NotFound, err)
		case *store.ErrNotValid:
			return nil, actions.NewError(actions.InvalidArgument, err)
		case *store.ErrPreconditionFailed:
			return nil, actions.NewError(actions.PreconditionFailed, err)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}

	return resource, nil
}

func (h Handlers) patchV3Resource(ctx context.Context, body []byte, name, namespace string, patcher patch.Patcher, conditions *store.ETagCondition) (interface{}, error) {
	payload := reflect.New(reflect.TypeOf(h.V3Resource).Elem())
	if err := json.Unmarshal(body, payload.Interface()); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}
	resource, ok := payload.Interface().(corev3.Resource)
	if !ok {
		return nil, actions.NewErrorf(actions.InvalidArgument)
	}

	req := storev2.NewResourceRequestFromResource(resource)
	req.Namespace = namespace
	req.Name = name

	w, err := wrap.ResourceWithoutValidation(resource)
	if err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	if err := h.StoreV2.Patch(ctx, req, w, patcher, conditions); err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, actions.NewError(actions.NotFound, err)
		case *store.ErrNotValid:
			return nil, actions.NewError(actions.InvalidArgument, err)
		case *store.ErrPreconditionFailed:
			return nil, actions.NewError(actions.PreconditionFailed, err)
		default:
			return nil, actions.NewError(actions.InternalErr, err)
		}
	}

	// Unwrap the updated resource
	resource, err = w.Unwrap()
	if err != nil {
		return nil, actions.NewError(actions.InternalErr, err)
	}

	return resource, nil
}

func validatePatch(data []byte, vars map[string]string) error {
	type body struct {
		Metadata *corev2.ObjectMeta `json:"metadata"`
	}

	b := &body{}

	if err := json.Unmarshal(data, b); err != nil {
		return err
	}

	if b.Metadata == nil {
		return nil
	}

	namespace, err := url.PathUnescape(vars["namespace"])
	if err != nil {
		return err
	}
	if b.Metadata.Namespace != "" && b.Metadata.Namespace != namespace {
		return fmt.Errorf(
			"the namespace of the resource (%s) does not match the namespace in the URI (%s)",
			b.Metadata.Namespace,
			namespace,
		)
	}

	name, err := url.PathUnescape(vars["id"])
	if err != nil {
		return err
	}
	if b.Metadata.Name != "" && b.Metadata.Name != name {
		return fmt.Errorf(
			"the name of the resource (%s) does not match the name in the URI (%s)",
			b.Metadata.Name,
			name,
		)
	}

	return nil
}
