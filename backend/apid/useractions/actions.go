package useractions

import "golang.org/x/net/context"

// Fetcher define standard actions for exposing ability for viewer to query
// resources.
type Fetcher interface {
	Find(context.Context, QueryParams) (interface{}, error)
	Query(context.Context, QueryParams) ([]interface{}, error)
}

// Destroyer exposes actions for viewer to delete resources.
type Destroyer interface {
	Destroy(context.Context, QueryParams) error
}

// QueryParams keys inwhich to filter results of query.
type QueryParams map[string]string
