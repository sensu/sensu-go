package e2e

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sensu/sensu-go/types"
)

// test the create command with all the resources
func TestCreateCommand(t *testing.T) {
	t.Parallel()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(t)
	defer cleanup()

	tests := []struct {
		Name    string
		Fixture types.Resource
	}{
		{
			Name:    "Asset",
			Fixture: types.FixtureAsset("foo"),
		},
		{
			Name:    "Check",
			Fixture: types.FixtureCheck("foo"),
		},
		{
			Name:    "Entity",
			Fixture: types.FixtureEntity("foo"),
		},
		{
			Name:    "Environment",
			Fixture: types.FixtureEnvironment("foo"),
		},
		{
			Name:    "Event",
			Fixture: types.FixtureEvent("foo", "bar"),
		},
		{
			Name:    "Extension",
			Fixture: types.FixtureExtension("foo"),
		},
		{
			Name:    "EventFilter",
			Fixture: types.FixtureEventFilter("foo"),
		},
		{
			Name:    "Handler",
			Fixture: types.FixtureHandler("foo"),
		},
		{
			Name:    "Hook",
			Fixture: types.FixtureHook("foo"),
		},
		{
			Name:    "Mutator",
			Fixture: types.FixtureMutator("foo"),
		},
		{
			Name:    "Organization",
			Fixture: types.FixtureOrganization("foo"),
		},
		{
			Name:    "Role",
			Fixture: types.FixtureRole("foo", "bar", "baz"),
		},
		{
			Name:    "Silenced",
			Fixture: types.FixtureSilenced("foo:bar"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			wrapper := types.Wrapper{
				Type:  test.Name,
				Value: test.Fixture,
			}
			b, err := json.Marshal(wrapper)
			if err != nil {
				t.Fatal(err)
			}
			sensuctl.stdin = bytes.NewReader(b)
			out, err := sensuctl.run("create")
			if err != nil {
				t.Fatal(err, string(out))
			}
		})
	}
}
