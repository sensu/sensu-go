package service

// Service ...TODO...
type Service struct{}

// New returns new instance of Service
func New() *Service {
	return &Service{}
}

// RegisterType registers a GraphQL type with the service.
func (service *Service) RegisterType(t Type, handler interface{}) {
	// Determine type
	// ...?
}
