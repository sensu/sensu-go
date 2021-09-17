package messaging

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

// MultiplexSignal creates a signal handler for the provided signal, and publishes
// all received instances of the signal to wizard bus. The topic it will be published
// to is the topic that would be returned for the signal by SignalTopic(sig).
//
// If the signal is already handled elsewhere, it will stop being handled there,
// as MultiplexSignal will call signal.Reset(sig) before calling signal.Notify.
func MultiplexSignal(ctx context.Context, bus MessageBus, sig os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Reset(sig)
	signal.Notify(ch, sig)
	topic := SignalTopic(sig)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case sig := <-ch:
				_ = bus.Publish(topic, sig)
			}
		}
	}()
}

// SignalTopic returns the topic name for the given signal. It can be used to
// subscribe to particular signals that have been multiplexed.
func SignalTopic(sig os.Signal) string {
	return fmt.Sprintf("signal:%s", sig.String())
}
