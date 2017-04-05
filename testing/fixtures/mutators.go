package fixtures

import "github.com/sensu/sensu-go/types"

var (
	mutatorFixtures = []*types.Mutator{
		&types.Mutator{
			Name:    "mutator1",
			Command: "cat",
			Timeout: 10,
		},
	}
)
