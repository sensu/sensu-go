package graphql

import (
	"testing"

	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/testutil"
)

func TestValidateMaxDepth(t *testing.T) {
	testutil.ExpectPassesRule(t, ProvideMaxDepthRule, `
	query MyQuery {
		forwardTo(cluster: "~") {
			forwardTo(cluster: "~") {
				health {
					postgresql {
						healthy
					}
				}
			}
		}
	}`)

	// NOTE: the graphql-go testutil requires a line and column in the query document
	// 	of the error node. The query is formatted this way to easily find the `healthy`
	// 	node which breaks the max depth rule
	testutil.ExpectFailsRule(t, ProvideMaxDepthRule, `query MyQuery {forwardTo(cluster: "~") {forwardTo(cluster: "~") {forwardTo(cluster: "~") {forwardTo(cluster: "~") {health {postgresql {
healthy # <-- this node is at depth six, breaking the max depth of 5 rule
		}}}}}}}`,
		[]gqlerrors.FormattedError{testutil.RuleError("Max depth exceeded", 2, 1)})
}
