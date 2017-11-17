package graphqlschema

import (
	"testing"

	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
)

type CheckMutationSuite struct {
	suite.Suite
	mutationSuite
}

func (t *CheckMutationSuite) SetupTest() {
	// Ensure query is run with viwer that has full access.
	t.populateContext(contextWithFullAccess)

	// Ensure the store has default org / env
	t.populateStore(func(ctx context.Context, st store.Store) {
		st.UpdateOrganization(ctx, types.FixtureOrganization("default"))
		st.UpdateEnvironment(ctx, types.FixtureEnvironment("default"))
	})
}

func (t *CheckMutationSuite) TestCreateSuccess() {
	result, errs := t.runQuery(
		t.T(),
		`
			mutation myMutation($inputs: CreateCheckInput!) {
				createCheck(input: $inputs) {
					check {
						name
					}
				}
			}
		`,
		map[string]interface{}{
			"inputs": map[string]interface{}{
				"clientMutationId": "1",
				"name":             "my-check",
				"organization":     "default",
				"environment":      "default",
				"interval":         30,
			},
		},
	)

	t.Empty(errs)
	t.Equal("my-check", result.Get("createCheck", "check", "name"))
}

func (t *CheckMutationSuite) TestUpdateSuccess() {
	t.populateStore(func(ctx context.Context, st store.Store) {
		st.UpdateCheckConfig(ctx, types.FixtureCheckConfig("my-check"))
	})

	result, errs := t.runQuery(
		t.T(),
		`
			mutation myMutation($inputs: UpdateCheckInput!) {
				updateCheck(input: $inputs) {
					check {
						name
						interval
					}
				}
			}
		`,
		map[string]interface{}{
			"inputs": map[string]interface{}{
				"clientMutationId": "1",
				"name":             "my-check",
				"organization":     "default",
				"environment":      "default",
				"interval":         30,
			},
		},
	)

	t.Empty(errs)
	t.Equal(30, result.Get("updateCheck", "check", "interval"))
}

func (t *CheckMutationSuite) TestDestroySuccess() {
	check := types.FixtureCheckConfig("my-check")
	gid := globalid.CheckTranslator.EncodeToString(check)

	t.populateStore(func(ctx context.Context, st store.Store) {
		st.UpdateCheckConfig(ctx, check)
	})

	result, errs := t.runQuery(
		t.T(),
		`
			mutation myMutation($inputs: DestroyCheckInput!) {
				destroyCheck(input: $inputs) {
					destroyedCheckId
				}
			}
		`,
		map[string]interface{}{
			"inputs": map[string]interface{}{
				"clientMutationId": "1",
				"id":               gid,
			},
		},
	)

	t.Empty(errs)
	t.Equal(gid, result.Get("destroyCheck", "destroyedCheckId"))
}

func (t *CheckMutationSuite) TestCreateInvalidInputs() {
	result, errs := t.runQuery(
		t.T(),
		`
			mutation myMutation($inputs: CreateCheckInput!) {
				createCheck(input: $inputs) {
					check {
						name
					}
				}
			}
		`,
		map[string]interface{}{
			"inputs": map[string]interface{}{
				"clientMutationId": "1",
				"name":             "my-check",
				"organization":     "badorg",
				"environment":      "very-bad-env",
				"interval":         0,
			},
		},
	)

	t.NotEmpty(errs)
	t.Nil(result.Get("createCheck"))
}

func (t *CheckMutationSuite) TestUpdateInvalidInputs() {
	result, errs := t.runQuery(
		t.T(),
		`
			mutation myMutation($inputs: UpdateCheckInput!) {
				updateCheck(input: $inputs) {
					check {
						name
					}
				}
			}
		`,
		map[string]interface{}{
			"inputs": map[string]interface{}{
				"clientMutationId": "1",
				"name":             "my-check",
				"organization":     "badorg",
				"environment":      "very-bad-env",
				"interval":         0,
			},
		},
	)

	t.NotEmpty(errs)
	t.Nil(result.Get("updateCheck"))
}

func TestCheckMutations(testRunner *testing.T) {
	runSuites(
		testRunner,
		new(CheckMutationSuite),
	)
}
