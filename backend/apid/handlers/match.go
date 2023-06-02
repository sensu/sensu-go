package handlers

import (
	"context"
	"net/http"

	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

const (
	ifMatchHdr     = "If-Match"
	ifNoneMatchHdr = "If-None-Match"
)

func matchHeaderContext(req *http.Request) (context.Context, error) {
	ctx := req.Context()
	if value := req.Header.Get(ifMatchHeader); value != "" {
		ifMatch, err := storev2.ReadIfMatch(value)
		if err != nil {
			return nil, err
		}
		ctx = storev2.ContextWithIfMatch(ctx, ifMatch)
	}
	if value := req.Header.Get(ifNoneMatchHeader); value != "" {
		ifNoneMatch, err := storev2.ReadIfNoneMatch(value)
		if err != nil {
			return nil, err
		}
		ctx = storev2.ContextWithIfNoneMatch(ctx, ifNoneMatch)
	}

	return ctx, nil
}
