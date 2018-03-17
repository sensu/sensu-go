package types

// Resource represents a Sensu resource.
type Resource interface {
	// URIPath gives the path to the resource, e.g. /checks/checkname
	URIPath() string

	// Validate checks if the fields in the resource are valid.
	Validate() error
}
