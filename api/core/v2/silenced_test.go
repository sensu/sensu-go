package v2

import (
	"reflect"
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

func TestSilencedMatches(t *testing.T) {
	testCases := []struct {
		name         string
		silenced     *Silenced
		check        string
		subscription string
		expected     bool
	}{
		{
			name:         "nil silence",
			silenced:     nil,
			check:        "",
			subscription: "",
			expected:     false,
		},
		{
			name:         "No subscription or check",
			silenced:     FixtureSilenced("*:*"),
			subscription: "matches",
			check:        "anything",
			expected:     true,
		},
		{
			name:         "Subscription matches, no check",
			silenced:     FixtureSilenced("foo:*"),
			subscription: "foo",
			check:        "wildcard",
			expected:     true,
		},
		{
			name:         "Subscription matches, check doesn't",
			silenced:     FixtureSilenced("foo:bar"),
			subscription: "foo",
			check:        "nomatch",
			expected:     false,
		},
		{
			name:         "Check matches, no subscription",
			silenced:     FixtureSilenced("*:foo"),
			subscription: "anything",
			check:        "foo",
			expected:     true,
		},
		{
			name:         "Check matches, subscription doesn't",
			silenced:     FixtureSilenced("foo:bar"),
			subscription: "nomatch",
			check:        "bar",
			expected:     false,
		},
		{
			name:         "Both check and subscription match",
			silenced:     FixtureSilenced("foo:bar"),
			subscription: "foo",
			check:        "bar",
			expected:     true,
		},
		{
			name:         "empty subscription is the same as wildcard",
			silenced:     &Silenced{Subscription: "", Check: "foo"},
			subscription: "",
			check:        "foo",
			expected:     true,
		},
		{
			name:         "empty check is the same as wildcard",
			silenced:     &Silenced{Subscription: "foo", Check: ""},
			subscription: "foo",
			check:        "",
			expected:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.silenced.Matches(tc.check, tc.subscription))
		})
	}
}

func TestSilencedFields(t *testing.T) {
	tests := []struct {
		name    string
		args    Resource
		wantKey string
		want    string
	}{
		{
			name:    "exposes name",
			args:    FixtureSilenced("a:b"),
			wantKey: "silenced.name",
			want:    "a:b",
		},
		{
			name:    "exposes expire_on_resolve",
			args:    FixtureSilenced("a:b"),
			wantKey: "silenced.expire_on_resolve",
			want:    "false",
		},
		{
			name: "exposes labels",
			args: &Silenced{
				ObjectMeta: ObjectMeta{
					Labels: map[string]string{"region": "philadelphia"},
				},
			},
			wantKey: "silenced.labels.region",
			want:    "philadelphia",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SilencedFields(tt.args)
			if !reflect.DeepEqual(got[tt.wantKey], tt.want) {
				t.Errorf("SilencedFields() = got[%s] %v, want[%s] %v", tt.wantKey, got[tt.wantKey], tt.wantKey, tt.want)
			}
		})
	}
}
