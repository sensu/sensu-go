package v2

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureSilenced(t *testing.T) {
	s := FixtureSilenced("test_subscription:test_check")
	s.Expire = 60
	s.ExpireOnResolve = true
	s.Creator = "creator@example.com"
	s.Reason = "test reason"
	s.Namespace = "default"
	assert.NotNil(t, s)
	assert.NotNil(t, s.Name)
	assert.Equal(t, "test_subscription:test_check", s.Name)
	assert.NotNil(t, s.Expire)
	assert.NotNil(t, s.ExpireOnResolve)
	assert.NotNil(t, s.Expire)
	assert.NotNil(t, s.Creator)
	assert.NotNil(t, s.Check)
	assert.NotNil(t, s.Reason)
	assert.NotNil(t, s.Subscription)
	assert.NotNil(t, s.Namespace)

	s = FixtureSilenced("entity:test_subscription:test_check")
	assert.Equal(t, "entity:test_subscription", s.Subscription)
	assert.Equal(t, "test_check", s.Check)
}

// Validation should fail when we don't provide a CheckName or Subscription
func TestSilencedValidate(t *testing.T) {
	var s Silenced
	assert.Error(t, s.Validate())
}

func TestSortSilencedByID(t *testing.T) {
	a := FixtureSilenced("Abernathy:*")
	b := FixtureSilenced("Bernard:*")
	c := FixtureSilenced("Clementine:*")
	d := FixtureSilenced("Dolores:*")

	testCases := []struct {
		name      string
		inRecords []*Silenced
		expected  []*Silenced
	}{
		{
			name:      "d, c, b, a",
			inRecords: []*Silenced{d, c, b, a},
			expected:  []*Silenced{a, b, c, d},
		},
		{
			name:      "c, d, a, b",
			inRecords: []*Silenced{c, d, a, b},
			expected:  []*Silenced{a, b, c, d},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(SortSilencedByName(tc.inRecords))
			assert.EqualValues(t, tc.expected, tc.inRecords)
		})
	}
}

func TestSortSilencedByBegin(t *testing.T) {
	a := FixtureSilenced("Abernathy:*")
	a.Begin = 5
	b := FixtureSilenced("Bernard:*")
	b.Begin = 10
	c := FixtureSilenced("Clementine:*")
	c.Begin = 50

	in := []*Silenced{b, a, c}
	sort.Sort(SortSilencedByBegin(in))
	assert.EqualValues(t, []*Silenced{a, b, c}, in)
}
