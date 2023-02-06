package backend

import (
	"context"
	"time"

	"github.com/sensu/sensu-go/backend/store"
)

func CheckInLoop(ctx context.Context, backendName string, opc store.OperatorConcierge) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	state := store.OperatorState{
		Type:           store.BackendOperator,
		Name:           backendName,
		CheckInTimeout: 10 * time.Second,
		Present:        true,
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := opc.CheckIn(ctx, state); err != nil {
				logger.WithError(err).Error("error checking-in backend operator")
			}
		}
	}
}
