package graphql

import (
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestMutationTypeExecuteCheck(t *testing.T) {
	inputs := schema.ExecuteCheckInput{}
	params := schema.MutationExecuteCheckFieldResolverParams{}
	params.Args.Input = &inputs

	impl := mutationsImpl{checkExecutor: mockCheckExecutor{}}

	// Success
	body, err := impl.ExecuteCheck(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	impl.checkExecutor = mockCheckExecutor{err: errors.New("wow")}
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
	impl := mutationsImpl{}

	// Success
	mock := struct {
		checkFinder
		checkReplacer
	}{
		mockCheckFinder{record: check},
		mockCheckReplacer{},
	}

	impl.checkReplacer = mock
	body, err := impl.UpdateCheck(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure - no check
	mock = struct {
		checkFinder
		checkReplacer
	}{
		mockCheckFinder{err: errors.New("test")},
		mockCheckReplacer{},
	}
	impl.checkReplacer = mock

	body, err = impl.UpdateCheck(params)
	assert.Error(t, err)
	assert.Empty(t, body)

	// Failure - replace fails
	mock = struct {
		checkFinder
		checkReplacer
	}{
		mockCheckFinder{record: check},
		mockCheckReplacer{err: errors.New("test")},
	}
	impl.checkReplacer = mock

	body, err = impl.UpdateCheck(params)
	assert.Error(t, err)
	assert.Empty(t, body)
}

func TestMutationTypeDeleteEntityField(t *testing.T) {
	inputs := schema.DeleteRecordInput{}
	params := schema.MutationDeleteEntityFieldResolverParams{}
	params.Args.Input = &inputs

	// Success
	impl := mutationsImpl{}
	impl.entityDestroyer = mockEntityDestroyer{}
	body, err := impl.DeleteEntity(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	impl.entityDestroyer = mockEntityDestroyer{err: errors.New("wow")}
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

	// Success
	impl := mutationsImpl{}
	impl.eventDestroyer = mockEventDestroyer{}
	body, err := impl.DeleteEvent(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Bad gid
	params.Args.Input = &schema.DeleteRecordInput{ID: "tests"}
	body, err = impl.DeleteEvent(params)
	assert.Error(t, err)
	assert.Nil(t, body)

	// Destroy failed
	impl.eventDestroyer = mockEventDestroyer{err: errors.New("test")}
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

	// Success
	impl := mutationsImpl{}
	impl.silenceCreator = mockSilenceCreator{}
	body, err := impl.CreateSilence(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	impl.silenceCreator = mockSilenceCreator{err: errors.New("wow")}
	body, err = impl.CreateSilence(params)
	assert.Error(t, err)
	assert.Nil(t, body)
}

func TestMutationTypeDeleteSilenceField(t *testing.T) {
	inputs := schema.DeleteRecordInput{}
	params := schema.MutationDeleteSilenceFieldResolverParams{}
	params.Args.Input = &inputs

	// Success
	impl := mutationsImpl{}
	impl.silenceDestroyer = mockSilenceDestroyer{}
	body, err := impl.DeleteSilence(params)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	// Failure
	impl.silenceDestroyer = mockSilenceDestroyer{err: errors.New("wow")}
	body, err = impl.DeleteSilence(params)
	assert.Error(t, err)
	assert.Nil(t, body)
}
