package api

// ListOptions represents the various options that can be used when listing
// resources.
type ListOptions struct {
	FieldSelector string
	LabelSelector string

	// ContinueToken is the current pagination token.
	ContinueToken string

	// ChunkSize is the number of objects to fetch per page when taking
	// advantage of the API's pagination capabilities. ChunkSize <= 0 means
	// fetch everything all at once; do not use pagination.
	ChunkSize int
}
