package graphql

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	v2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/types"
)

var _ schema.MutationFieldResolvers = (*mutationsImpl)(nil)

//
// Implement MutationFieldResolvers
//

type mutationsImpl struct {
	factory ClientFactory
}

type deleteRecordPayload struct {
	schema.DeleteRecordPayloadAliases
}

//
// Implement generic PUT mutation
//

// PutWrapped implements response to request for the 'putWrapped' field.
func (r *mutationsImpl) PutWrapped(p schema.MutationPutWrappedFieldResolverParams) (interface{}, error) {
	var ret types.Wrapper
	var err error

	raw := p.Args.Raw
	upsert := p.Args.Upsert

	// decode given
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()
	if err = dec.Decode(&ret); err != nil {
		return map[string]interface{}{
			"errors": wrapInputErrors("raw", err),
		}, nil
	}

	// Determine if we should send the wrapped resource or only the resource
	// itself
	var bytes []byte
	if ret.APIVersion == "core/v2" {
		bytes, err = json.Marshal(ret.Value)
	} else {
		bytes, err = json.Marshal(ret)
	}
	if err != nil {
		return map[string]interface{}{
			"errors": wrapInputErrors("raw", err),
		}, nil
	}

	path := ret.Value.URIPath()
	client := r.factory.NewWithContext(p.Context)
	if upsert {
		err = client.Put(path, bytes)
	} else {
		// Note: Minor Hack
		// -----------------------------------------------------------------
		// If the `upsert` parameter on this mutation is `false`, we want to
		// return an error if the resource already exists, instead of
		// updating the existing resource.
		//
		// To do this, we make a request to the POST endpoint instead of the
		// PUT endpoint. However, since the URI path resolved from
		// `ret.Value.URIPath()` includes the name of the resource in the
		// path string, we need to remove the name in order to call the
		// correct endpoint.
		//
		// Ex.
		// POST `/api/core/v2/namespaces/default/checks`
		// PUT `/api/core/v2/namespaces/default/checks/check-cpu`
		//
		// Here we slice off the resource name in order to call the correct
		// endpoint. This works for the resource we currently create from the
		// Web UI (Checks and Handlers), but it's possible that this MAY not
		// work for future resources.
		suffix := "/" + url.PathEscape(ret.Value.GetObjectMeta().Name)
		pathSubstr := strings.TrimSuffix(path, suffix)
		err = client.Post(pathSubstr, bytes)
	}
	if err != nil {
		return map[string]interface{}{
			"errors": wrapInputErrors("raw", err),
		}, nil
	}

	return map[string]interface{}{
		"node":   ret.Value,
		"errors": []stdErr{},
	}, nil
}

//
// Implement check mutations
//

// CreateCheck implements response to request for the 'createCheck' field.
func (r *mutationsImpl) CreateCheck(p schema.MutationCreateCheckFieldResolverParams) (interface{}, error) {
	inputs := p.Args.Input

	var check types.CheckConfig
	check.Name = inputs.Name
	check.Namespace = inputs.Namespace

	rawArgs := p.ResolveParams.Args
	if err := copyCheckInputs(&check, rawArgs["input"]); err != nil {
		return nil, err
	}

	ctx := contextWithNamespace(p.Context, inputs.Namespace)
	client := r.factory.NewWithContext(ctx)

	err := client.CreateCheck(&check)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"check":            &check,
	}, nil
}

// UpdateCheck implements response to request for the 'updateCheck' field.
func (r *mutationsImpl) UpdateCheck(p schema.MutationUpdateCheckFieldResolverParams) (interface{}, error) {
	components, _ := globalid.Decode(p.Args.Input.ID)
	ctx := setContextFromComponents(p.Context, components)

	client := r.factory.NewWithContext(ctx)
	check, err := client.FetchCheck(components.UniqueComponent())
	if err != nil {
		return nil, err
	}

	rawArgs := p.ResolveParams.Args
	if err := copyCheckInputs(check, rawArgs["input"]); err != nil {
		return nil, err
	}

	err = client.UpdateCheck(check)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"check":            check,
	}, nil
}

// DeleteCheck implements response to request for the 'deleteCheck' field.
func (r *mutationsImpl) DeleteCheck(p schema.MutationDeleteCheckFieldResolverParams) (interface{}, error) {
	components, _ := globalid.Decode(p.Args.Input.ID)
	ctx := setContextFromComponents(p.Context, components)

	client := r.factory.NewWithContext(ctx)

	err := client.DeleteCheck(components.Namespace(), components.UniqueComponent())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"deletedId":        components.String(),
	}, nil
}

// ExecuteCheck implements response to request for the 'executeCheck' field.
func (r *mutationsImpl) ExecuteCheck(p schema.MutationExecuteCheckFieldResolverParams) (interface{}, error) {
	components, _ := globalid.Decode(p.Args.Input.ID)
	ctx := setContextFromComponents(p.Context, components)

	client := r.factory.NewWithContext(ctx)
	check, err := client.FetchCheck(components.UniqueComponent())
	if err != nil {
		return nil, err
	}

	adhocReq := types.AdhocRequest{
		ObjectMeta:    check.ObjectMeta,
		Subscriptions: p.Args.Input.Subscriptions,
		Reason:        p.Args.Input.Reason,
	}
	err = client.ExecuteCheck(&adhocReq)
	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"errors":           wrapInputErrors("id", err),
	}, nil
}

func copyCheckInputs(r *types.CheckConfig, args interface{}) error {
	input, ok := args.(map[string]interface{})
	if !ok {
		return errors.New("given unexpected arguments")
	}
	props, ok := input["props"].(map[string]interface{})
	if !ok {
		return errors.New("given unexpected arguments; props is not a map")
	}

	return mapstructure.Decode(props, r)
}

type checkMutationPayload struct {
	schema.CreateCheckPayloadAliases
}

//
// Implement entity mutations
//

// DeleteEntity implements response to request for the 'deleteEntity' field.
func (r *mutationsImpl) DeleteEntity(p schema.MutationDeleteEntityFieldResolverParams) (interface{}, error) {
	components, _ := globalid.Decode(p.Args.Input.ID)
	ctx := setContextFromComponents(p.Context, components)

	client := r.factory.NewWithContext(ctx)

	err := client.DeleteEntity(components.Namespace(), components.UniqueComponent())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"deletedId":        p.Args.Input.ID,
	}, nil
}

//
// Implement event mutations
//

// ResolveEvent implements response to request for the 'resolveEvent' field.
func (r *mutationsImpl) ResolveEvent(p schema.MutationResolveEventFieldResolverParams) (interface{}, error) {
	components, err := decodeEventGID(p.Args.Input.ID)
	if err != nil {
		return nil, err
	}

	ctx := setContextFromComponents(p.Context, components)
	client := r.factory.NewWithContext(ctx)

	event, err := client.FetchEvent(components.EntityName(), components.CheckName())
	if err != nil {
		return nil, err
	}

	if event.HasCheck() && event.Check.Status > 0 {
		event.Check.Status = 0
		event.Check.Output = "Resolved manually with " + p.Args.Input.Source
		event.Timestamp = int64(time.Now().Unix())

		err = client.UpdateEvent(event)
		if err != nil {
			return nil, err
		}
	}

	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"event":            event,
	}, nil
}

// DeleteEvent implements response to request for the 'deleteEvent' field.
func (r *mutationsImpl) DeleteEvent(p schema.MutationDeleteEventFieldResolverParams) (interface{}, error) {
	components, err := decodeEventGID(p.Args.Input.ID)
	if err != nil {
		return nil, err
	}

	ctx := setContextFromComponents(p.Context, components)
	client := r.factory.NewWithContext(ctx)

	err = client.DeleteEvent(components.Namespace(), components.EntityName(), components.CheckName())
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"deletedId":        p.Args.Input.ID,
	}, nil
}

func decodeEventGID(gid string) (globalid.EventComponents, error) {
	components := globalid.EventComponents{}
	parsedComponents, err := globalid.Parse(gid)
	if err != nil {
		return components, err
	}

	if parsedComponents.Resource() != globalid.EventTranslator.ForResourceNamed() {
		return components, errors.New("given id does not appear to reference event")
	}

	components = globalid.NewEventComponents(parsedComponents)
	return components, nil
}

//
// Implement event filter mutations
//

// DeleteEventFilter implements response to request for the 'deleteEventFilter' field.
func (r *mutationsImpl) DeleteEventFilter(p schema.MutationDeleteEventFilterFieldResolverParams) (interface{}, error) {
	components, _ := globalid.Decode(p.Args.Input.ID)
	if components.Resource() != v2.EventFiltersResource {
		return nil, errors.New("given ID must be a event filter")
	}

	ctx := setContextFromComponents(p.Context, components)
	client := r.factory.NewWithContext(ctx)

	err := client.DeleteFilter(components.Namespace(), components.UniqueComponent())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"deletedId":        components.String(),
	}, nil
}

//
// Implement handler mutations
//

// DeleteHandler implements response to request for the 'deleteHandler' field.
func (r *mutationsImpl) DeleteHandler(p schema.MutationDeleteHandlerFieldResolverParams) (interface{}, error) {
	components, _ := globalid.Decode(p.Args.Input.ID)
	ctx := setContextFromComponents(p.Context, components)

	client := r.factory.NewWithContext(ctx)

	err := client.DeleteHandler(components.Namespace(), components.UniqueComponent())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"deletedId":        components.String(),
	}, nil
}

//
// Implement mutator mutations
//

// DeleteMutator implements response to request for the 'deleteMutator' field.
func (r *mutationsImpl) DeleteMutator(p schema.MutationDeleteMutatorFieldResolverParams) (interface{}, error) {
	components, _ := globalid.Decode(p.Args.Input.ID)
	if components.Resource() != v2.MutatorsResource {
		return nil, errors.New("given ID must be a mutator")
	}

	ctx := setContextFromComponents(p.Context, components)
	client := r.factory.NewWithContext(ctx)

	err := client.DeleteMutator(components.Namespace(), components.UniqueComponent())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"deletedId":        components.String(),
	}, nil
}

//
// Implement silenced mutations
//

// CreateSilence implements response to request for the 'createSilence' field.
func (r *mutationsImpl) CreateSilence(p schema.MutationCreateSilenceFieldResolverParams) (interface{}, error) {
	inputs := p.Args.Input

	var silence types.Silenced
	silence.Check = inputs.Check
	silence.Subscription = inputs.Subscription
	silence.Namespace = inputs.Namespace
	copySilenceInputs(&silence, inputs.Props)

	ctx := contextWithNamespace(p.Context, inputs.Namespace)
	client := r.factory.NewWithContext(ctx)

	err := client.CreateSilenced(&silence)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"clientMutationId": inputs.ClientMutationID,
		"silence":          &silence,
	}, nil
}

// DeleteSilence implements response to request for the 'deleteSilence' field.
func (r *mutationsImpl) DeleteSilence(p schema.MutationDeleteSilenceFieldResolverParams) (interface{}, error) {
	components, _ := globalid.Parse(p.Args.Input.ID)
	ctx := setContextFromComponents(p.Context, components)

	client := r.factory.NewWithContext(ctx)
	err := client.DeleteSilenced(components.Namespace(), components.UniqueComponent())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"deletedId":        p.Args.Input.ID,
	}, nil
}

func copySilenceInputs(r *types.Silenced, ins *schema.SilenceInputs) {
	r.Begin = 0
	if ins.Begin.After(time.Now()) {
		r.Begin = ins.Begin.Unix()
	}

	r.Reason = ins.Reason
	r.Expire = int64(ins.Expire)
	r.ExpireOnResolve = ins.ExpireOnResolve
}
