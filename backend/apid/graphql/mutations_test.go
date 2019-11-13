package graphql

import (
	"context"
	"errors"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMutationTypePutWrappedUpsertTrue(t *testing.T) {
	params := schema.MutationPutWrappedFieldResolverParams{}
	params.Args.Raw = `
		{
			"type": "Silenced",
			"metadata": {
				"namespace":"sensu-devel",
				"name": "test:fred"
			},
			"spec": {
				"check": "fred",
				"creator": "asdfasdf"
			}
		}
	`
	params.Args.Upsert = true

	client := new(MockGenericClient)
	client.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	client.On("SetTypeMeta", mock.Anything).Return(nil)
	cfg := ServiceConfig{GenericClient: client}
	impl := mutationsImpl{svc: cfg}

	// Success
	body, err := impl.PutWrapped(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Bad JSON
	params.Args.Raw = `{ "type.... ]`
	body, err = impl.PutWrapped(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	client.On("Put", mock.Anything, mock.Anything).Return(errors.New("test")).Once()
	body, err = impl.PutWrapped(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)
}

func TestMutationTypePutWrappedUpsertFalse(t *testing.T) {
	params := schema.MutationPutWrappedFieldResolverParams{}
	params.Args.Raw = `
		{
			"type": "Silenced",
			"metadata": {
				"namespace":"sensu-devel",
				"name": "test:fred"
			},
			"spec": {
				"check": "fred",
				"creator": "asdfasdf"
			}
		}
	`
	params.Args.Upsert = false

	client := new(MockGenericClient)
	cfg := ServiceConfig{GenericClient: client}
	client.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
	client.On("SetTypeMeta", mock.Anything).Return(nil)
	impl := mutationsImpl{svc: cfg}

	// Success
	body, err := impl.PutWrapped(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Bad JSON
	params.Args.Raw = `{ "type.... ]`
	body, err = impl.PutWrapped(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	client.On("Create", mock.Anything, mock.Anything).Return(errors.New("test")).Once()
	body, err = impl.PutWrapped(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)
}

func TestMutationTypeExecuteCheck(t *testing.T) {
	inputs := schema.ExecuteCheckInput{}
	params := schema.MutationExecuteCheckFieldResolverParams{}
	params.Args.Input = &inputs

	check := corev2.FixtureCheckConfig("test")
	client := new(MockCheckClient)
	client.On("FetchCheck", mock.Anything, mock.Anything).Return(check, nil)
	client.On("ExecuteCheck", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	cfg := ServiceConfig{CheckClient: client}
	impl := mutationsImpl{svc: cfg}

	// Success
	body, err := impl.ExecuteCheck(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	client.On("ExecuteCheck", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test")).Once()
	body, err = impl.ExecuteCheck(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)
}

func TestMutationTypeUpdateCheck(t *testing.T) {
	inputs := schema.UpdateCheckInput{}
	params := schema.MutationUpdateCheckFieldResolverParams{}
	params.Args.Input = &inputs
	params.ResolveParams.Args = map[string]interface{}{
		"input": map[string]interface{}{
			"props": map[string]interface{}{
				"command": "yes",
			},
		},
	}

	check := corev2.FixtureCheckConfig("a")
	client := new(MockCheckClient)
	client.On("FetchCheck", mock.Anything, mock.Anything).Return(check, nil).Once()
	client.On("UpdateCheck", mock.Anything, mock.Anything).Return(nil).Once()
	cfg := ServiceConfig{CheckClient: client}

	// Success
	impl := mutationsImpl{svc: cfg}
	body, err := impl.UpdateCheck(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure - no check
	client.On("FetchCheck", mock.Anything, mock.Anything).Return(check, errors.New("404")).Once()
	body, err = impl.UpdateCheck(params)
	assert.Error(t, err)
	assert.Empty(t, body)

	// Failure - replace fails
	client.On("FetchCheck", mock.Anything, mock.Anything).Return(check, nil).Once()
	client.On("UpdateCheck", mock.Anything, mock.Anything).Return(errors.New("fail")).Once()
	body, err = impl.UpdateCheck(params)
	assert.Error(t, err)
	assert.Empty(t, body)
}

func TestMutationTypeDeleteEntityField(t *testing.T) {
	inputs := schema.DeleteRecordInput{}
	params := schema.MutationDeleteEntityFieldResolverParams{}
	params.Args.Input = &inputs

	entity := corev2.FixtureEntity("abc")
	client := new(MockEntityClient)
	cfg := ServiceConfig{EntityClient: client}
	impl := mutationsImpl{svc: cfg}

	// Success
	client.On("FetchEntity", mock.Anything, mock.Anything).Return(entity, nil)
	client.On("DeleteEntity", mock.Anything, mock.Anything).Return(nil).Once()
	body, err := impl.DeleteEntity(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	client.On("DeleteEntity", mock.Anything, mock.Anything).Return(errors.New("fail")).Once()
	body, err = impl.DeleteEntity(params)
	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestMutationTypeDeleteEventField(t *testing.T) {
	evt := corev2.FixtureEvent("a", "b")
	gid := globalid.EventTranslator.EncodeToString(context.Background(), evt)

	inputs := schema.DeleteRecordInput{ID: gid}
	params := schema.MutationDeleteEventFieldResolverParams{}
	params.Args.Input = &inputs

	client := new(MockEventClient)
	cfg := ServiceConfig{EventClient: client}
	impl := mutationsImpl{svc: cfg}

	// Success
	client.On("DeleteEvent", mock.Anything, "a", "b").Return(nil).Once()
	body, err := impl.DeleteEvent(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Bad gid
	params.Args.Input = &schema.DeleteRecordInput{ID: "tests"}
	body, err = impl.DeleteEvent(params)
	assert.Error(t, err)
	assert.Nil(t, body)

	// Destroy failed
	client.On("DeleteEvent", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("err")).Once()
	body, err = impl.DeleteEvent(params)
	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestMutationTypeDeleteHandlerField(t *testing.T) {
	hd := corev2.FixtureHandler("a")
	gid := globalid.HandlerTranslator.EncodeToString(context.Background(), hd)

	inputs := schema.DeleteRecordInput{ID: gid}
	params := schema.MutationDeleteHandlerFieldResolverParams{}
	params.Args.Input = &inputs

	client := new(MockHandlerClient)
	cfg := ServiceConfig{HandlerClient: client}
	impl := mutationsImpl{svc: cfg}

	// Success
	client.On("DeleteHandler", mock.Anything, "a").Return(nil).Once()
	body, err := impl.DeleteHandler(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	client.On("DeleteHandler", mock.Anything, mock.Anything).Return(errors.New("err")).Once()
	body, err = impl.DeleteHandler(params)
	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestMutationTypeDeleteMutatorField(t *testing.T) {
	mut := corev2.FixtureMutator("a")
	gid := globalid.MutatorTranslator.EncodeToString(context.Background(), mut)

	inputs := schema.DeleteRecordInput{ID: gid}
	params := schema.MutationDeleteMutatorFieldResolverParams{}
	params.Args.Input = &inputs

	client := new(MockMutatorClient)
	cfg := ServiceConfig{MutatorClient: client}
	impl := mutationsImpl{svc: cfg}

	// Success
	client.On("DeleteMutator", mock.Anything, "a").Return(nil).Once()
	body, err := impl.DeleteMutator(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	client.On("DeleteMutator", mock.Anything, mock.Anything).Return(errors.New("err")).Once()
	body, err = impl.DeleteMutator(params)
	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestMutationTypeDeleteEventFilterField(t *testing.T) {
	flr := corev2.FixtureEventFilter("a")
	gid := globalid.EventFilterTranslator.EncodeToString(context.Background(), flr)

	inputs := schema.DeleteRecordInput{ID: gid}
	params := schema.MutationDeleteEventFilterFieldResolverParams{}
	params.Args.Input = &inputs

	client := new(MockEventFilterClient)
	cfg := ServiceConfig{EventFilterClient: client}
	impl := mutationsImpl{svc: cfg}

	// Success
	client.On("DeleteEventFilter", mock.Anything, mock.Anything).Return(nil).Once()
	body, err := impl.DeleteEventFilter(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	client.On("DeleteEventFilter", mock.Anything, mock.Anything).Return(errors.New("err")).Once()
	body, err = impl.DeleteEventFilter(params)
	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestMutationTypeCreateSilenceField(t *testing.T) {
	inputs := schema.CreateSilenceInput{
		Namespace: "a",
		Props:     &schema.SilenceInputs{},
	}
	params := schema.MutationCreateSilenceFieldResolverParams{}
	params.Args.Input = &inputs

	client := new(MockSilencedClient)
	cfg := ServiceConfig{SilencedClient: client}
	impl := mutationsImpl{svc: cfg}

	// Success
	client.On("UpdateSilenced", mock.Anything, mock.Anything).Return(nil).Once()
	body, err := impl.CreateSilence(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	client.On("UpdateSilenced", mock.Anything, mock.Anything).Return(errors.New("test")).Once()
	body, err = impl.CreateSilence(params)
	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestMutationTypeDeleteSilenceField(t *testing.T) {
	inputs := schema.DeleteRecordInput{}
	params := schema.MutationDeleteSilenceFieldResolverParams{}
	params.Args.Input = &inputs

	client := new(MockSilencedClient)
	cfg := ServiceConfig{SilencedClient: client}
	impl := mutationsImpl{svc: cfg}

	// Success
	client.On("DeleteSilencedByName", mock.Anything, mock.Anything).Return(nil).Once()
	body, err := impl.DeleteSilence(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	client.On("DeleteSilencedByName", mock.Anything, mock.Anything).Return(errors.New("test")).Once()
	body, err = impl.DeleteSilence(params)
	assert.Error(t, err)
	assert.Nil(t, body)
}
