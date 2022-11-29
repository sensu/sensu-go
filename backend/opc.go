package backend

import (
	"context"
	"os"
	"time"

	"github.com/sensu/sensu-go/backend/store"
)

func CheckInLoop(ctx context.Context, opc store.OperatorConcierge) {
	backendName, err := os.Hostname()
	if err != nil {
		// According to `man gethostname`, this should never happen, unless
		// there is a bug in Go's use of gethostname
		panic(err)
	}
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
