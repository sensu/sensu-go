package v2

import (
	"bytes"
	"context"
	"fmt"
	"strings"
)

// contextKeyTxInfo is the context key that identifies a TxInfo.
type contextKeyTxInfo struct{}

// TxInfo is a record of store actions that were performed as a result of a
// ResourceRequest.
type TxInfo struct {
	// Records will contain an entry for each store record that was affected
	// by the resource request.
	Records []TxRecordInfo
}

// TxRecordInfo is a record of a store write.
type TxRecordInfo struct {
	Created  bool
	Updated  bool
	Deleted  bool
	ETag     ETag
	PrevETag ETag
}

// ContextWithTxInfo returns a new context that contains the supplied TxInfo.
// Implementations that read *TxInfo from a context bearing it should write
// transaction stats into it that reflect the effects of the transaction.
func ContextWithTxInfo(ctx context.Context, tx *TxInfo) context.Context {
	return context.WithValue(ctx, contextKeyTxInfo{}, tx)
}

// TxInfoFromContext returns the *TxInfo from the context, or nil if it is missing.
func TxInfoFromContext(ctx context.Context) *TxInfo {
	val := ctx.Value(contextKeyTxInfo{})
	if val == nil {
		return nil
	}
	return val.(*TxInfo)
}

// IfMatch is a list of Etags
type IfMatch []ETag
type contextKeyIfMatch struct{}

func (m IfMatch) Matches(etag ETag) bool {
	for _, candidate := range m {
		if bytes.Equal(candidate, etag) {
			return true
		}
	}
	return false
}

func (m IfMatch) String() string {
	parts := make([]string, 0, len(m))
	for _, v := range m {
		parts = append(parts, fmt.Sprintf("%q", v.String()))
	}
	return strings.Join(parts, ", ")
}

func ReadIfMatch(header string) (IfMatch, error) {
	var result IfMatch
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)[1 : len(part)-1]
		etag, err := DecodeETag(part)
		if err != nil {
			continue
		}
		result = append(result, etag)
	}
	return result, nil
}

// ContextWithIfMatch returns a new context that contains the supplied IfMatch.
func ContextWithIfMatch(ctx context.Context, list IfMatch) context.Context {
	return context.WithValue(ctx, contextKeyIfMatch{}, list)
}

// IfMatch returns the IfMatch from the context, or nil if it is missing.
func IfMatchFromContext(ctx context.Context) IfMatch {
	val := ctx.Value(contextKeyIfMatch{})
	if val == nil {
		return nil
	}
	return val.(IfMatch)
}

// IfNoneMatch is a list of ETags.
type IfNoneMatch []ETag
type contextKeyIfNoneMatch struct{}

// ContextWithIfNoneMatch returns a new context that contains the supplied IfNoneMatch.
func ContextWithIfNoneMatch(ctx context.Context, list IfNoneMatch) context.Context {
	return context.WithValue(ctx, contextKeyIfNoneMatch{}, list)
}

// IfMatch returns the IfNoneMatch from the context, or nil if it is missing.
func IfNoneMatchFromContext(ctx context.Context) IfNoneMatch {
	val := ctx.Value(contextKeyIfNoneMatch{})
	if val == nil {
		return nil
	}
	return val.(IfNoneMatch)
}

func (m IfNoneMatch) String() string {
	parts := make([]string, 0, len(m))
	for _, v := range m {
		parts = append(parts, fmt.Sprintf("%q", v.String()))
	}
	return strings.Join(parts, ", ")
}

func ReadIfNoneMatch(header string) (IfNoneMatch, error) {
	var result IfNoneMatch
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)[1 : len(part)-1]
		etag, err := DecodeETag(part)
		if err != nil {
			continue
		}
		result = append(result, etag)
	}
	return result, nil
}

func (m IfNoneMatch) Matches(etag ETag) bool {
	for _, candidate := range m {
		if bytes.Equal(candidate, etag) {
			return false
		}
	}
	return true
}
