package mockring

import (
	"path"

	types "github.com/sensu/sensu-go/types"
)

// RingGetter ...
type RingGetter map[string]types.Ring

// GetRing ...
func (g RingGetter) GetRing(p ...string) types.Ring {
	s := path.Join(p...)
	return g[s]
}
