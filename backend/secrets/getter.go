package secrets

import (
	"context"
)

// Getter represents an abstracted secret getter.
type Getter interface {
	// Get gets the name of the provider and secret ID associated with the Sensu secret name.
	Get(ctx context.Context, name string) (provider string, id string, err error)
}
