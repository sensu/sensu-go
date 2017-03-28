package fixtures

import "github.com/sensu/sensu-go/types"

var (
	checkFixtures = []*types.Check{
		&types.Check{
			Name:        "check1",
			Interval:    60,
			Subscribers: []string{"subscriber1"},
			Command:     "command1",
		},
	}
)
