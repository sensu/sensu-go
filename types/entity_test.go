package types

import (
	"encoding/json"
	"sort"
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

	// Invalid namespace
	assert.Error(t, e.Validate())
	e.Namespace = "default"

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

	// Find exported field
	val, err := e.Get("ID")
	require.NoError(t, err)
	assert.EqualValues(t, "myAgent", val)
}

func TestEntityUnmarshal(t *testing.T) {
	entity := Entity{}

	// Unmarshal
	err := json.Unmarshal([]byte(`{"id": "myAgent"}`), &entity)
	require.NoError(t, err)

	// Existing exported fields were properly set
	assert.Equal(t, "myAgent", entity.ID)
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
