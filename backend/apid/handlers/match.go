package handlers

import (
	"context"
	"encoding/base64"
	"net/http"

	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

const (
	ifMatchHdr     = "If-Match"
	ifNoneMatchHdr = "If-None-Match"
)

func matchHeaderContext(req *http.Request) context.Context {
	var (
		ifMatch     storev2.IfMatch
		ifNoneMatch storev2.IfNoneMatch
	)
	ctx := req.Context()
	ifMatches := req.Header.Values(ifMatchHdr)
	for _, value := range ifMatches {
		b, err := base64.RawStdEncoding.DecodeString(value)
		if err == nil {
			// no error reporting for bad headers, it's not worth it
			ifMatch = append(ifMatch, storev2.ETag(b))
		}
	}
	ifNoneMatches := req.Header.Values(ifNoneMatchHdr)
	for _, value := range ifNoneMatches {
		b, err := base64.RawStdEncoding.DecodeString(value)
		if err == nil {
			// no error reporting for bad headers, it's not worth it
			ifNoneMatch = append(ifNoneMatch, storev2.ETag(b))
		}
	}

	ctx = storev2.ContextWithIfMatch(ctx, ifMatch)
	ctx = storev2.ContextWithIfNoneMatch(ctx, ifNoneMatch)

	return ctx
}
