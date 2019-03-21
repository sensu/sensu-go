package graphql

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	client "github.com/sensu/sensu-go/backend/apid/graphql/mockclient"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMutationTypeExecuteCheck(t *testing.T) {
	inputs := schema.ExecuteCheckInput{}
	params := schema.MutationExecuteCheckFieldResolverParams{}
	params.Args.Input = &inputs

	check := types.FixtureCheckConfig("test")
	client, factory := client.NewClientFactory()
	client.On("FetchCheck", mock.Anything).Return(check, nil)
	client.On("ExecuteCheck", mock.Anything).Return(nil).Once()
	impl := mutationsImpl{factory: factory}

	// Success
	body, err := impl.ExecuteCheck(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	client.On("ExecuteCheck", mock.Anything).Return(errors.New("test")).Once()
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

	check := types.FixtureCheckConfig("a")
	client, factory := client.NewClientFactory()
	client.On("FetchCheck", mock.Anything).Return(check, nil).Once()
	client.On("UpdateCheck", mock.Anything).Return(nil).Once()

	// Success
	impl := mutationsImpl{factory: factory}
	body, err := impl.UpdateCheck(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure - no check
	client.On("FetchCheck", mock.Anything).Return(check, errors.New("404")).Once()
	body, err = impl.UpdateCheck(params)
	assert.Error(t, err)
	assert.Empty(t, body)

	// Failure - replace fails
	client.On("FetchCheck", mock.Anything).Return(check, nil).Once()
	client.On("UpdateCheck", mock.Anything).Return(errors.New("fail")).Once()
	body, err = impl.UpdateCheck(params)
	assert.Error(t, err)
	assert.Empty(t, body)
}

func TestMutationTypeDeleteEntityField(t *testing.T) {
	inputs := schema.DeleteRecordInput{}
	params := schema.MutationDeleteEntityFieldResolverParams{}
	params.Args.Input = &inputs

	entity := types.FixtureEntity("abc")
	client, factory := client.NewClientFactory()
	impl := mutationsImpl{factory: factory}

	// Success
	client.On("FetchEntity", mock.Anything).Return(entity, nil)
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
	evt := types.FixtureEvent("a", "b")
	gid := globalid.EventTranslator.EncodeToString(evt)

	inputs := schema.DeleteRecordInput{ID: gid}
	params := schema.MutationDeleteEventFieldResolverParams{}
	params.Args.Input = &inputs

	client, factory := client.NewClientFactory()
	impl := mutationsImpl{factory: factory}

	// Success
	client.On("DeleteEvent", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
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

func TestMutationTypeCreateSilenceField(t *testing.T) {
	inputs := schema.CreateSilenceInput{
		Namespace: "a",
		Props:     &schema.SilenceInputs{},
	}
	params := schema.MutationCreateSilenceFieldResolverParams{}
	params.Args.Input = &inputs

	client, factory := client.NewClientFactory()
	impl := mutationsImpl{factory: factory}

	// Success
	client.On("CreateSilenced", mock.Anything).Return(nil).Once()
	body, err := impl.CreateSilence(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	client.On("CreateSilenced", mock.Anything).Return(errors.New("test")).Once()
	body, err = impl.CreateSilence(params)
	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestMutationTypeDeleteSilenceField(t *testing.T) {
	inputs := schema.DeleteRecordInput{}
	params := schema.MutationDeleteSilenceFieldResolverParams{}
	params.Args.Input = &inputs

	client, factory := client.NewClientFactory()
	impl := mutationsImpl{factory: factory}

	// Success
	client.On("DeleteSilenced", mock.Anything, mock.Anything).Return(nil).Once()
	body, err := impl.DeleteSilence(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	client.On("DeleteSilenced", mock.Anything, mock.Anything).Return(errors.New("test")).Once()
	body, err = impl.DeleteSilence(params)
	assert.Error(t, err)
	assert.Nil(t, body)
}
