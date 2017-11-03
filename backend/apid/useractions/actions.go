package useractions

// Fetcher define standard actions for exposing ability for viewer to query
// resources.
type Fetcher interface {
	Find(params QueryParams) (interface{}, error)
	Query(params QueryParams) ([]interface{}, error)
}

// Destroyer exposes actions for viewer to delete resources.
type Destroyer interface {
	Destroy(params QueryParams) error
}

// QueryParams keys inwhich to filter results of query.
type QueryParams map[string]string
