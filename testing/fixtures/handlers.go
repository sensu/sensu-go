package fixtures

import "github.com/sensu/sensu-go/types"

var (
	handlerFixtures = []*types.Handler{
		&types.Handler{
			Name: "handler1",
			Type: "pipe",
			Pipe: types.HandlerPipe{
				Command: "cat",
				Timeout: 10,
			},
		},
	}
)
