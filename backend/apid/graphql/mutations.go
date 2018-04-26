package graphql

import (
	"errors"
	"time"

	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/graphql"
	"github.com/sensu/sensu-go/types"
)

var _ schema.MutationFieldResolvers = (*mutationsImpl)(nil)

//
// Implement HookFieldResolvers
//

type mutationsImpl struct {
	checkController actions.CheckController
	eventController actions.EventController
}

func newMutationImpl(store store.Store, getter types.QueueGetter, bus messaging.MessageBus) *mutationsImpl {
	return &mutationsImpl{
		checkController: actions.NewCheckController(store, getter),
		eventController: actions.NewEventController(store, bus),
	}
}

type deleteRecordPayload struct {
	schema.DeleteRecordPayloadAliases
}

//
// Implement check mutations
//

// CreateCheck implements response to request for the 'createCheck' field.
func (r *mutationsImpl) CreateCheck(p schema.MutationCreateCheckFieldResolverParams) (interface{}, error) {
	inputs := p.Args.Input

	var check types.CheckConfig
	check.Name = inputs.Name
	check.Organization = inputs.Ns.Organization
	check.Environment = inputs.Ns.Environment
	copyCheckInputs(&check, inputs.Props)

	err := r.checkController.Create(p.Context, check)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"check":            check,
	}, nil
}

// UpdateCheck implements response to request for the 'updateCheck' field.
func (r *mutationsImpl) UpdateCheck(p schema.MutationUpdateCheckFieldResolverParams) (interface{}, error) {
	inputs := p.Args.Input
	components, _ := globalid.Decode(inputs.ID.(string))

	var check types.CheckConfig
	check.Name = components.UniqueComponent()
	check.Organization = components.Organization()
	check.Environment = components.Environment()
	copyCheckInputs(&check, inputs.Props)

	err := r.checkController.Update(p.Context, check)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"clientMutationId": inputs.ClientMutationID,
		"check":            check,
	}, nil
}

// DeleteCheck implements response to request for the 'deleteCheck' field.
func (r *mutationsImpl) DeleteCheck(p schema.MutationDeleteCheckFieldResolverParams) (interface{}, error) {
	components, _ := globalid.Decode(p.Args.Input.ID.(string))
	ctx := setContextFromComponents(p.Context, components)

	err := r.checkController.Destroy(ctx, components.UniqueComponent())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"deletedId":        components.String(),
	}, nil
}

func copyCheckInputs(r *types.CheckConfig, ins *schema.CheckConfigInputs) {
	r.RuntimeAssets = ins.Assets
	r.Command = ins.Command
	r.Handlers = ins.Handlers
	r.Interval = uint32(ins.Interval)
	if ins.HighFlapThreshold >= 0 {
		r.HighFlapThreshold = &types.FlapThreshold{Value: uint32(ins.HighFlapThreshold)}
	} else {
		r.HighFlapThreshold = nil
	}
	if ins.LowFlapThreshold >= 0 {
		r.LowFlapThreshold = &types.FlapThreshold{Value: uint32(ins.LowFlapThreshold)}
	} else {
		r.HighFlapThreshold = nil
	}
	r.Subscriptions = ins.Subscriptions
	r.Publish = ins.Publish
}

type checkMutationPayload struct {
	schema.CreateCheckPayloadAliases
}

// IsTypeOf is used to determine if a given value is associated with the type
func (*mutationsImpl) IsTypeOf(s interface{}, p graphql.IsTypeOfParams) bool {
	return false
}

//
// Implement event mutations
//

// ResolveEvent implements response to request for the 'resolveEvent' field.
func (r *mutationsImpl) ResolveEvent(p schema.MutationResolveEventFieldResolverParams) (interface{}, error) {
	components, _ := globalid.Decode(p.Args.Input.ID.(string))
	ctx := setContextFromComponents(p.Context, components)

	evComponents, ok := components.(globalid.EventComponents)
	if !ok {
		return nil, errors.New("given id does not appear to reference event")
	}

	event, err := r.eventController.Find(ctx, evComponents.EntityName(), evComponents.CheckName())
	if err != nil {
		return nil, err
	}

	if event.Check != nil && event.Check.Status > 0 {
		event.Check.Status = 0
		event.Check.Output = "Resolved manually with " + p.Args.Input.Source
		event.Timestamp = int64(time.Now().Unix())

		err = r.eventController.CreateOrReplace(ctx, *event)
		if err != nil {
			return nil, err
		}
	}

	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"event":            event,
	}, nil
}
