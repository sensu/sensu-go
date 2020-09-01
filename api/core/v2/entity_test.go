package v2

import (
	"encoding/json"
	"reflect"
	"sort"
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

func TestSortEntitiesByID(t *testing.T) {
	a := FixtureEntity("Abernathy")
	b := FixtureEntity("Bernard")
	c := FixtureEntity("Clementine")
	d := FixtureEntity("Dolores")

	testCases := []struct {
		name      string
		inDir     bool
		inRecords []*Entity
		expected  []*Entity
	}{
		{
			name:      "Sorts ascending",
			inDir:     true,
			inRecords: []*Entity{d, c, a, b},
			expected:  []*Entity{a, b, c, d},
		},
		{
			name:      "Sorts descending",
			inDir:     false,
			inRecords: []*Entity{d, a, c, b},
			expected:  []*Entity{d, c, b, a},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(SortEntitiesByID(tc.inRecords, tc.inDir))
			assert.EqualValues(t, tc.expected, tc.inRecords)
		})
	}
}

func TestSortEntitiesByLastSeen(t *testing.T) {
	a := FixtureEntity("Abernathy")
	a.LastSeen = 4
	b := FixtureEntity("Bernard")
	b.LastSeen = 3
	c := FixtureEntity("Clementine")
	c.LastSeen = 2
	d := FixtureEntity("Dolores")
	d.LastSeen = 1

	testCases := []struct {
		name      string
		inDir     bool
		inRecords []*Entity
		expected  []*Entity
	}{
		{
			name:      "Sort by last seen",
			inDir:     false,
			inRecords: []*Entity{d, a, c, b},
			expected:  []*Entity{a, b, c, d},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(SortEntitiesByLastSeen(tc.inRecords))
			assert.EqualValues(t, tc.expected, tc.inRecords)
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
