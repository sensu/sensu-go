package graphql

import (
	"context"
	"errors"
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

type mockEventDestroyer struct {
	err error
}

func (m mockEventDestroyer) Destroy(ctx context.Context, a, b string) error {
	return m.err
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
