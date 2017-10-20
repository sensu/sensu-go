package relay

import (
	"testing"

	"github.com/sensu/sensu-go/graphql/globalid"
	"github.com/stretchr/testify/assert"
)

func TestRegisterLookup(t *testing.T) {
	checkResolver := NodeResolver{Translator: globalid.CheckTranslator}
	event1Resolver := NodeResolver{
		Translator: globalid.EventTranslator,
		IsKindOf:   func(globalid.Components) bool { return true },
	}
	event2Resolver := NodeResolver{
		Translator: globalid.EventTranslator,
		IsKindOf:   func(globalid.Components) bool { return false },
	}

	register := NodeRegister{}
	register.RegisterResolver(checkResolver)
	register.RegisterResolver(event1Resolver)
	register.RegisterResolver(event2Resolver)

	testCases := []struct {
		gid  string
		want *NodeResolver
	}{
		{"srn:doggo:stbernard", nil},
		{"srn:checks:my-check", &checkResolver},
		{"srn:events:my-event", &event1Resolver},
	}

	for _, tc := range testCases {
		t.Run(tc.gid, func(t *testing.T) {
			components, _ := globalid.Parse(tc.gid)
			resolver := register.Lookup(components)
			if tc.want == nil {
				assert.Nil(t, resolver)
			} else {
				assert.NotNil(t, resolver)
			}
		})
	}
}
