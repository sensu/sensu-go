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
		&types.Handler{
			Name:     "handler2",
			Type:     "set",
			Handlers: []string{"handler1", "unknown"},
		},
		&types.Handler{
			Name:     "handler3",
			Type:     "set",
			Handlers: []string{"handler1", "handler2"},
		},
		&types.Handler{
			Name:     "handler4",
			Type:     "set",
			Handlers: []string{"handler2", "handler3"},
		},
	}
)
