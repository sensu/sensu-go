package schedulerd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

var (
	adhocQueueName = "adhocRequest"
)

// QueueStore contains the methods necessary for creating a queue.
type QueueStore interface {
	store.Store
	queue.Get
}

// AdhocRequestExecutor takes new check requests from the adhoc queue and runs
// them.
type AdhocRequestExecutor struct {
	adhocQueue queue.Interface
	store      store.Store
	bus        messaging.MessageBus
}

// NewAdhocRequestExecutor returns a new AdhocRequestExecutor.
func NewAdhocRequestExecutor(store QueueStore, bus messaging.MessageBus) *AdhocRequestExecutor {
	fmt.Println(store)
	executor := &AdhocRequestExecutor{
		adhocQueue: store.NewQueue(adhocQueueName),
		store:      store,
		bus:        bus,
	}
	return executor
}

// Start the dequeue function
func (a *AdhocRequestExecutor) Start(ctx context.Context) error {
	// this is probably a little redundant?
	return a.listenQueue(ctx)
}

func (a *AdhocRequestExecutor) listenQueue(ctx context.Context) error {
	var err error
	for {
		// listen to the queue, unmarshal value into a check request, and execute it
		item, err := a.adhocQueue.Dequeue(ctx)
		if err != nil {
			return err
		}
		var check types.CheckConfig
		if err := json.Unmarshal([]byte(item.Value), &check); err != nil {
			// if there is an error unmarshaling we're going to repeatedly nack this
			// item
			item.Nack(ctx)
			return err
		}
		if err := a.executeCheck(&check); err != nil {
			item.Nack(ctx)
			return err
		}
		err = item.Ack(ctx)
	}
	return err
}

// also needs to support proxy check requests, ignore splay settings
// don't need to worry about updating assets/hooks/etc because they come
// straight from the store
func (a *AdhocRequestExecutor) executeCheck(check *types.CheckConfig) error {
	request := &types.CheckRequest{}
	request.Config = check
	var err error
	for _, sub := range check.Subscriptions {
		topic := messaging.SubscriptionTopic(check.Organization, check.Environment, sub)
		logger.Debugf("sending check request for %s on topic %s", check.Name, topic)

		if pubErr := a.bus.Publish(topic, request); pubErr != nil {
			logger.Info("error publishing check request: ", pubErr.Error())
			err = pubErr
		}
	}
	return err
}
