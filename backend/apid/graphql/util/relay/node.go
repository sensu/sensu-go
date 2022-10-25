package util_relay

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/graphql/globalid"
	"github.com/sensu/sensu-go/backend/apid/graphql/relay"
	util_api "github.com/sensu/sensu-go/backend/apid/graphql/util/api"
	"github.com/sensu/sensu-go/types"
)

type Fetcher interface {
	SetTypeMeta(corev2.TypeMeta) error
	Get(context.Context, string, corev3.Resource) error
}

// MakeNodeResolver instatiates a new node resolver given a generic client and
// typemeta.
func MakeNodeResolver(client Fetcher, tm corev2.TypeMeta) func(relay.NodeResolverParams) (interface{}, error) {
	return func(p relay.NodeResolverParams) (interface{}, error) {
		if err := client.SetTypeMeta(tm); err != nil {
			return nil, err
		}
		raw, err := types.ResolveRaw(tm.APIVersion, tm.Type)
		if err != nil {
			return nil, err
		}
		// TODO(ccressent): I'm really not sure I can assume that raw can always
		// be assumed to meet the corev3.Resource interface...
		r := raw.(corev3.Resource)
		err = client.Get(p.Context, p.IDComponents.UniqueComponent(), r)
		return util_api.UnwrapGetResult(util_api.UnwrapResource(r), err)
	}
}

// ToGID produces a globalid for the given resource; an empty string is
// returned if the reverse lookup failed. Use ReverseLookup if you need to
// handle the error.
func ToGID(ctx context.Context, r interface{}) string {
	e, err := globalid.ReverseLookup(r)
	if err != nil {
		logger.WithError(err).Warn("toGID: unable to find type")
		return ""
	}
	return e.Encode(ctx, util_api.UnwrapResource(r)).String()
}
