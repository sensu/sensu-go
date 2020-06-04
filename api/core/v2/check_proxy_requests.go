package v2

import (
	"errors"

	"github.com/sensu/sensu-go/api/core/v2/internal/js"
)

// FixtureProxyRequests returns a fixture for a ProxyRequests object.
func FixtureProxyRequests(splay bool) *ProxyRequests {
	splayCoverage := uint32(0)
	if splay {
		splayCoverage = DefaultSplayCoverage
	}
	return &ProxyRequests{
		Splay:         splay,
		SplayCoverage: splayCoverage,
	}
}

// Validate returns an error if the ProxyRequests does not pass validation tests
func (p *ProxyRequests) Validate() error {
	if p.SplayCoverage > 100 {
		return errors.New("proxy request splay coverage must be between 0 and 100")
	}

	if (p.Splay) && (p.SplayCoverage == 0) {
		return errors.New("proxy request splay coverage must be greater than 0 if splay is enabled")
	}

	return js.ParseExpressions(p.EntityAttributes)
}
