package graphql

import (
	"testing"

	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/testutil"
)

// use graphql-go testutil to confirm validator provided correctly
func TestMaxDepthRule(t *testing.T) {
	testutil.ExpectPassesRule(t, MaxDepthRule(5), `
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

	testutil.ExpectFailsRule(t, MaxDepthRule(1), `
		query MyQuery {   # depth 0
			health {        # depth 1
				healthy				# depth 2 <-- exceeds max depth of 1
			}
		}`,
		// `healthy` node which breaks rule is on line 4, column 5 of the query document
		[]gqlerrors.FormattedError{testutil.RuleError("Max depth exceeded", 4, 5)})
}
