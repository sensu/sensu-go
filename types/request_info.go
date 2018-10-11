package types

// RequestInfo represents various information extracted from an HTTP request.
type RequestInfo struct {
	APIGroup     string
	APIVersion   string
	Namespace    string
	Resource     string
	ResourceName string
	Verb         string
}
