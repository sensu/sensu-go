package eventd

import (
	"context"
	"time"

	"github.com/sensu/sensu-go/backend/store"
)

func (e *Eventd) monitorCheckTTLs(ctx context.Context) {
	req := store.MonitorOperatorsRequest{
		Type:           store.CheckOperator,
		ControllerType: store.BackendOperator,
		ControllerName: e.backendName,
		Micromanage:    true,
		Every:          time.Second,
		ErrorHandler: func(err error) {
			logger.WithError(err).Error("error monitoring check TTLs")
		},
	}
	stateCh := e.operatorMonitor.MonitorOperators(ctx, req)
	for {
		select {
		case <-ctx.Done():
			return
		case states := <-stateCh:
			for _, state := range states {
				if err := e.handleCheckTTLNotification(ctx, state); err != nil {
					logger.WithError(err).Error("error handling check TTL failure")
				}
			}
		}
	}
}
