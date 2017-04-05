package fixtures

import "github.com/sensu/sensu-go/types"

var (
	handlerFixtures = []*types.Handler{
		&types.Handler{
			Name:    "handler1",
			Type:    "pipe",
			Mutator: "mutator1",
			Pipe: types.HandlerPipe{
				Command: "cat",
				Timeout: 10,
			},
		},
	}
)
