package queue_test

import (
	"context"
	"testing"

	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/queue/testspec"
)

func TestMemoryClient(t *testing.T) {
	q := queue.NewMemoryClient()

	ctx := context.Background()

	testspec.RunClientTestSuite(t, ctx, q)
}
