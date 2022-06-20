package v2

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityValidate(t *testing.T) {
	var e Entity

	// Invalid ID
	assert.Error(t, e.Validate())
	e.Name = "foo"

	// Invalid class
	assert.Error(t, e.Validate())
	e.EntityClass = "agent"

	// Invalid namespace
	assert.Error(t, e.Validate())
	e.Namespace = "default"

	// Valid entity
	assert.NoError(t, e.Validate())
}

func TestFixtureEntityIsValid(t *testing.T) {
	e := FixtureEntity("entity")
	assert.Equal(t, "entity", e.Name)
	assert.NoError(t, e.Validate())
}

func TestEntityUnmarshal(t *testing.T) {
	entity := Entity{}

	// Unmarshal
	err := json.Unmarshal([]byte(`{"metadata": {"name": "myAgent"}}`), &entity)
	require.NoError(t, err)

	// Existing exported fields were properly set
	assert.Equal(t, "myAgent", entity.Name)
}

func TestEntityMarshal(t *testing.T) {
	entity := FixtureEntity("myAgent")

	bytes, err := json.Marshal(entity)
	require.NoError(t, err)
	assert.Contains(t, string(bytes), "myAgent")
}

func TestEntityFields(t *testing.T) {
	tests := []struct {
		name    string
		args    Fielder
		wantKey string
		want    string
	}{
		{
			name:    "exposes name",
			args:    FixtureEntity("ap-007"),
			wantKey: "entity.name",
			want:    "ap-007",
		},
		{
			name:    "exposes deregister",
			args:    &Entity{Deregister: true},
			wantKey: "entity.deregister",
			want:    "true",
		},
		{
			name: "exposes labels",
			args: &Entity{
				ObjectMeta: ObjectMeta{
					Labels: map[string]string{"region": "philadelphia"},
				},
			},
			wantKey: "entity.labels.region",
			want:    "philadelphia",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.Fields()
			if !reflect.DeepEqual(got[tt.wantKey], tt.want) {
				t.Errorf("Entity.Fields() = got[%s] %v, want[%s] %v", tt.wantKey, got[tt.wantKey], tt.wantKey, tt.want)
			}
		})
	}
}

func TestAddEntitySubscription(t *testing.T) {
	tests := []struct {
		name          string
		entityName    string
		subscriptions []string
		want          []string
	}{
		{
			name:          "the entity subscription is added if missing",
			entityName:    "foo",
			subscriptions: []string{},
			want:          []string{"entity:foo"},
		},
		{
			name:          "the entity subscription is not added if already present",
			entityName:    "foo",
			subscriptions: []string{"entity:foo"},
			want:          []string{"entity:foo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddEntitySubscription(tt.entityName, tt.subscriptions); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addEntitySubscription() = %v, want %v", got, tt.want)
			}
		})
	}
}
