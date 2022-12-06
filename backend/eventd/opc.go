package eventd

import (
	"context"
	"os"
	"time"

	"github.com/sensu/sensu-go/backend/store"
)

var hostname string

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		panic(err)
	}
}

func (e *Eventd) monitorCheckTTLs(ctx context.Context) {
	req := store.MonitorOperatorsRequest{
		Type:           store.CheckOperator,
		ControllerType: store.BackendOperator,
		ControllerName: hostname,
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
