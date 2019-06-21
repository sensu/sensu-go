package v2

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureMutator(t *testing.T) {
	fixture := FixtureMutator("fixture")
	assert.Equal(t, "fixture", fixture.Name)
	assert.NoError(t, fixture.Validate())
}

func TestMutatorValidate(t *testing.T) {
	var m Mutator

	// Invalid name
	assert.Error(t, m.Validate())
	m.Name = "foo"

	// Invalid command
	assert.Error(t, m.Validate())
	m.Command = "echo 'foo'"

	// Invalid namespace
	assert.Error(t, m.Validate())
	m.Namespace = "default"

	// Valid mutator
	assert.NoError(t, m.Validate())
}

func TestSortMutatorsByName(t *testing.T) {
	a := FixtureMutator("Abernathy")
	b := FixtureMutator("Bernard")
	c := FixtureMutator("Clementine")
	d := FixtureMutator("Dolores")

	testCases := []struct {
		name     string
		inDir    bool
		inChecks []*Mutator
		expected []*Mutator
	}{
		{
			name:     "Sorts ascending",
			inDir:    true,
			inChecks: []*Mutator{d, c, a, b},
			expected: []*Mutator{a, b, c, d},
		},
		{
			name:     "Sorts descending",
			inDir:    false,
			inChecks: []*Mutator{d, a, c, b},
			expected: []*Mutator{d, c, b, a},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(SortMutatorsByName(tc.inChecks, tc.inDir))
			assert.EqualValues(t, tc.expected, tc.inChecks)
		})
	}
}
