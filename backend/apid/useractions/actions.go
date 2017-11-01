package useractions

// QueryActions define standard actions for exposing ability for viewer to query
// resources.
type QueryActions interface {
	Find(identifier string) (interface{}, error)
	Query(params QueryParams) ([]interface{}, error)
}

// QueryParams keys inwhich to filter results of query.
type QueryParams map[string]string
