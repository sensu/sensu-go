package graphql

import (
	"errors"
	"time"

	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var _ schema.MutationFieldResolvers = (*mutationsImpl)(nil)

//
// Implement MutationFieldResolvers
//

type mutationsImpl struct {
	checkCtrl     actions.CheckController
	checkExecutor checkExecutor

	entityDestroyer entityDestroyer

	eventFinder    eventFinder
	eventReplacer  eventReplacer
	eventDestroyer eventDestroyer

	silenceCreator   silenceCreator
	silenceDestroyer silenceDestroyer
}

func newMutationImpl(store store.Store, getter types.QueueGetter, bus messaging.MessageBus) *mutationsImpl {
	eventCtrl := actions.NewEventController(store, bus)
	checkCtrl := actions.NewCheckController(store, getter)
	entityCtrl := actions.NewEntityController(store)
	silenceCtrl := actions.NewSilencedController(store)

	return &mutationsImpl{
		checkCtrl:     checkCtrl,
		checkExecutor: checkCtrl,

		entityDestroyer: entityCtrl,

		eventFinder:    eventCtrl,
		eventReplacer:  eventCtrl,
		eventDestroyer: eventCtrl,

		silenceCreator:   silenceCtrl,
		silenceDestroyer: silenceCtrl,
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

	err := r.checkCtrl.Create(p.Context, check)
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
	components, _ := globalid.Decode(inputs.ID)

	var check types.CheckConfig
	check.Name = components.UniqueComponent()
	check.Organization = components.Organization()
	check.Environment = components.Environment()
	copyCheckInputs(&check, inputs.Props)

	err := r.checkCtrl.Update(p.Context, check)
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
	components, _ := globalid.Decode(p.Args.Input.ID)
	ctx := setContextFromComponents(p.Context, components)

	err := r.checkCtrl.Destroy(ctx, components.UniqueComponent())
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

	check := components.UniqueComponent()
	adhocReq := types.AdhocRequest{
		Name:          check,
		Subscriptions: p.Args.Input.Subscriptions,
		Reason:        p.Args.Input.Reason,
	}

	err := r.checkExecutor.QueueAdhocRequest(ctx, check, &adhocReq)
	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"errors":           wrapInputErrors("id", err),
	}, nil
}

func copyCheckInputs(r *types.CheckConfig, ins *schema.CheckConfigInputs) {
	r.RuntimeAssets = ins.Assets
	r.Command = ins.Command
	r.Handlers = ins.Handlers
	r.Interval = uint32(ins.Interval)
	r.HighFlapThreshold = uint32(ins.HighFlapThreshold)
	r.LowFlapThreshold = uint32(ins.LowFlapThreshold)
	r.Subscriptions = ins.Subscriptions
	r.Publish = ins.Publish
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

	err := r.entityDestroyer.Destroy(ctx, components.UniqueComponent())
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
	event, err := r.eventFinder.Find(ctx, components.EntityName(), components.CheckName())
	if err != nil {
		return nil, err
	}

	if event.HasCheck() && event.Check.Status > 0 {
		event.Check.Status = 0
		event.Check.Output = "Resolved manually with " + p.Args.Input.Source
		event.Timestamp = int64(time.Now().Unix())

		err = r.eventReplacer.CreateOrReplace(ctx, *event)
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
	err = r.eventDestroyer.Destroy(ctx, components.EntityName(), components.CheckName())
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
// Implement silenced mutations
//

// CreateSilence implements response to request for the 'createSilence' field.
func (r *mutationsImpl) CreateSilence(p schema.MutationCreateSilenceFieldResolverParams) (interface{}, error) {
	inputs := p.Args.Input

	var silence types.Silenced
	silence.Check = inputs.Check
	silence.Subscription = inputs.Subscription
	silence.Organization = inputs.Ns.Organization
	silence.Environment = inputs.Ns.Environment
	copySilenceInputs(&silence, inputs.Props)

	err := r.silenceCreator.Create(p.Context, &silence)
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

	err := r.silenceDestroyer.Destroy(ctx, components.UniqueComponent())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"clientMutationId": p.Args.Input.ClientMutationID,
		"deletedId":        p.Args.Input.ID,
	}, nil
}

func copySilenceInputs(r *types.Silenced, ins *schema.SilenceInputs) {
	r.Reason = ins.Reason
	r.Begin = ins.Begin.Unix()
	r.Expire = int64(ins.Expire)
	r.ExpireOnResolve = ins.ExpireOnResolve
}
