package schedulerd

import (
	"context"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
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
	messages   chan *roundRobinMessage
	ctx        context.Context
	ringGetter types.RingGetter
	bus        messaging.MessageBus
}

// newRoundRobinScheduler creates a new roundRobinScheduler.
//
// When the scheduler is created, it starts a goroutine that will stop when
// the provided context is cancelled.
func newRoundRobinScheduler(ctx context.Context, bus messaging.MessageBus, rg types.RingGetter) *roundRobinScheduler {
	sched := &roundRobinScheduler{
		messages:   make(chan *roundRobinMessage),
		ctx:        ctx,
		bus:        bus,
		ringGetter: rg,
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
	ctx, cancel := context.WithTimeout(r.ctx, timeout)
	defer cancel()
	ring := r.ringGetter.GetRing("subscription", msg.subscription)
	entityID, err := ring.Next(ctx)
	if err != nil {
		r.logError(err, entityID, msg.req.Config.Name)
		return
	}
	// GetEntitySubscription gets a subscription that maps directly to an
	// entity.
	sub := types.GetEntitySubscription(entityID)
	cfg := msg.req.Config
	topic := messaging.SubscriptionTopic(cfg.Organization, cfg.Environment, sub)
	if err := r.bus.Publish(topic, msg.req); err != nil {
		r.logError(err, entityID, msg.req.Config.Name)
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

// logError logs errors and adds agentID and checkName as fields.
func (r *roundRobinScheduler) logError(err error, agentID, checkName string) {
	logger.
		WithFields(logrus.Fields{"agent": agentID, "check": checkName}).
		WithError(err).
		Error("error publishing round robin check request")
}
