package schedulerd

import (
	"context"
	"sync"

	time "github.com/echlebek/timeproxy"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/types"
)

// roundRobinMessage is a combination of a check request and a subscription
type roundRobinMessage struct {
	req          *types.CheckRequest
	subscription string
	wg           sync.WaitGroup
}

// roundRobinScheduler is an appendage of CheckScheduler. It exists to handle
// round-robin execution for subscriptions that specify it.
type roundRobinScheduler struct {
	messages chan *roundRobinMessage
	ctx      context.Context
	bus      messaging.MessageBus
}

// newRoundRobinScheduler creates a new roundRobinScheduler.
//
// When the scheduler is created, it starts a goroutine that will stop when
// the provided context is cancelled.
func newRoundRobinScheduler(ctx context.Context, bus messaging.MessageBus) *roundRobinScheduler {
	sched := &roundRobinScheduler{
		messages: make(chan *roundRobinMessage),
		ctx:      ctx,
		bus:      bus,
	}
	go sched.loop()
	return sched
}

// loop handles scheduler events
func (r *roundRobinScheduler) loop() {
	for {
		select {
		case <-r.ctx.Done():
			close(r.messages)
			return
		case msg := <-r.messages:
			go r.execute(msg)
		}
	}
}

// execute executes a round robin check request
func (r *roundRobinScheduler) execute(msg *roundRobinMessage) {
	defer msg.wg.Done()
	timeout := time.Second * time.Duration(msg.req.Config.Interval)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := r.bus.PublishDirect(ctx, msg.subscription, msg.req); err != nil {
		r.logError(err, msg.req.Config.Name)
	}
}

// Schedule schedules a check request to run in a round robin ring.
// Schedule returns a sync.WaitGroup that the caller can wait on to know
// when scheduling is completed.
func (r *roundRobinScheduler) Schedule(msg *roundRobinMessage) (*sync.WaitGroup, error) {
	if err := r.ctx.Err(); err != nil {
		return nil, err
	}
	msg.wg.Add(1)
	r.messages <- msg
	return &msg.wg, nil
}

// logError logs errors and adds agentName and checkName as fields.
func (r *roundRobinScheduler) logError(err error, checkName string) {
	logger.WithField("check", checkName).WithError(err).Error("error publishing round robin check request")
}
