package client

import "github.com/sensu/sensu-go/types"

// APIClient client methods across the Sensu API
type APIClient interface {
	EventAPIClient
	CheckAPIClient
}

// EventAPIClient client methods for events
type EventAPIClient interface {
	ListEvents() ([]types.Event, error)
}

// CheckAPIClient client methods for checks
type CheckAPIClient interface {
	//  	CreateCheck(*types.Check) error
}
