package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityValidate(t *testing.T) {
	var e Entity

	// Invalid ID
	assert.Error(t, e.Validate())
	e.ID = "foo"

	// Invalid class
	assert.Error(t, e.Validate())
	e.Class = "agent"

	// Invalid organization
	assert.Error(t, e.Validate())
	e.Organization = "default"

	// Invalid environment
	assert.Error(t, e.Validate())
	e.Environment = "default"

	// Valid entity
	assert.NoError(t, e.Validate())
}

func TestFixtureEntityIsValid(t *testing.T) {
	e := FixtureEntity("entity")
	assert.Equal(t, "entity", e.ID)
	assert.NoError(t, e.Validate())
}

func TestEntityGet(t *testing.T) {
	e := FixtureEntity("myAgent")
	e.SetExtendedAttributes([]byte(`{"foo": "bar"}`))

	// Find extended attr
	val, err := e.Get("foo")
	require.NoError(t, err)
	assert.EqualValues(t, "bar", val)

	// Find exported field
	val, err = e.Get("ID")
	require.NoError(t, err)
	assert.EqualValues(t, "myAgent", val)
}

func TestEntityUnmarshal(t *testing.T) {
	entity := Entity{}

	// Unmarshal
	err := json.Unmarshal([]byte(`{"id": "myAgent", "foo": "bar"}`), &entity)
	require.NoError(t, err)

	// Existing exported fields were properly set
	assert.Equal(t, "myAgent", entity.ID)

	// ExtendedAttribute
	f, err := entity.Get("foo")
	require.NoError(t, err)
	assert.EqualValues(t, "bar", f)
}

func TestEntityMarshal(t *testing.T) {
	entity := FixtureEntity("myAgent")
	entity.SetExtendedAttributes([]byte(`{"foo": "bar"}`))

	bytes, err := json.Marshal(entity)
	require.NoError(t, err)
	assert.Contains(t, string(bytes), "myAgent")
	assert.Contains(t, string(bytes), "foo")
	assert.Contains(t, string(bytes), "bar")
}
