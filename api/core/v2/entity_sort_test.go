package v2

import (
	"sort"
	testing "testing"

	"github.com/stretchr/testify/assert"
)

func TestSortEntitiesByID(t *testing.T) {
	a := FixtureEntity("Abernathy")
	b := FixtureEntity("Bernard")
	c := FixtureEntity("Clementine")
	d := FixtureEntity("Dolores")

	testCases := []struct {
		name      string
		asc       bool
		inRecords []*Entity
		expected  []*Entity
	}{
		{
			name:      "Sorts ascending",
			asc:       true,
			inRecords: []*Entity{d, c, a, b},
			expected:  []*Entity{a, b, c, d},
		},
		{
			name:      "Sorts descending",
			asc:       false,
			inRecords: []*Entity{d, a, c, b},
			expected:  []*Entity{d, c, b, a},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(SortEntitiesByID(tc.inRecords, tc.asc))
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
		asc       bool
		inRecords []*Entity
		expected  []*Entity
	}{
		{
			name:      "Sorts descending",
			asc:       false,
			inRecords: []*Entity{d, a, c, b},
			expected:  []*Entity{a, b, c, d},
		},
		{
			name:      "Sorts ascending",
			asc:       true,
			inRecords: []*Entity{d, a, c, b},
			expected:  []*Entity{d, c, b, a},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(SortEntitiesByLastSeen(tc.inRecords, tc.asc))
			assert.EqualValues(t, tc.expected, tc.inRecords)
		})
	}
}
