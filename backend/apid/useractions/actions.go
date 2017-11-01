package useractions

// Fetcher define standard actions for exposing ability for viewer to query
// resources.
type Fetcher interface {
	Find(params QueryParams) (interface{}, error)
	Query(params QueryParams) ([]interface{}, error)
}

// QueryParams keys inwhich to filter results of query.
type QueryParams map[string]string
