package v2

import (
	"reflect"
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

func TestMutatorFields(t *testing.T) {
	tests := []struct {
		name    string
		args    Resource
		wantKey string
		want    string
	}{
		{
			name:    "exposes name",
			args:    FixtureMutator("ninja-turtle"),
			wantKey: "mutator.name",
			want:    "ninja-turtle",
		},
		{
			name: "exposes labels",
			args: &Mutator{
				ObjectMeta: ObjectMeta{
					Labels: map[string]string{"region": "philadelphia"},
				},
			},
			wantKey: "mutator.labels.region",
			want:    "philadelphia",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MutatorFields(tt.args)
			if !reflect.DeepEqual(got[tt.wantKey], tt.want) {
				t.Errorf("MutatorFields() = got[%s] %v, want[%s] %v", tt.wantKey, got[tt.wantKey], tt.wantKey, tt.want)
			}
		})
	}
}

func TestValidateMutatorTypes(t *testing.T) {
	passTests := []string{"", "javascript", "pipe"}
	failTests := []string{"Javascript", "js", "Pipe"}
	for _, test := range passTests {
		mutator := FixtureMutator("foo")
		mutator.Type = test
		if mutator.Type == JavascriptMutator {
			mutator.Command = ""
			mutator.Eval = "return 'asdf';"
		}
		if err := mutator.Validate(); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range failTests {
		mutator := FixtureMutator("foo")
		mutator.Type = test
		if err := mutator.Validate(); err == nil {
			t.Fatal("expecte non-nil error")
		}
	}
}

func TestValidateMutatorCommandWithJavascript(t *testing.T) {
	mutator := FixtureMutator("foo")
	mutator.Command = "asdfasdf"
	mutator.Type = JavascriptMutator
	if err := mutator.Validate(); err == nil {
		t.Fatal("expected non-nil error")
	}
}
