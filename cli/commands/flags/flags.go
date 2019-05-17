package flags

const (
	// AllNamespaces is used to query all resources regardless of their namespace
	AllNamespaces = "all-namespaces"

	// Format is used to specify the expected output of the command
	Format = "format"

	// Interactive is used to specify if cli should be interactive
	Interactive = "interactive"

	// FieldSelector is used to provide a field expression that will be used as
	// a filter, typically when listing resources.
	FieldSelector = "field-selector"

	// LabelSelector is used to provide a label expression that will be used as
	// a filter, typically when listing resources.
	LabelSelector = "label-selector"

	// ChunkSize is used to specify that a list of objects is to be fetched in
	// chunks of the given size, using the API's pagination capabilities.
	ChunkSize = "chunk-size"
)
