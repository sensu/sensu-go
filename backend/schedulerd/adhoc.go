package schedulerd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/types"
)

var (
	adhocQueueName = "adhocRequest"
)

// AdhocRequestExecutor takes new check requests from the adhoc queue and runs
// them.
type AdhocRequestExecutor struct {
	adhocQueue     queue.Interface
	store          StateManagerStore
	bus            messaging.MessageBus
	ctx            context.Context
	cancel         context.CancelFunc
	listenQueueErr chan error
}

// NewAdhocRequestExecutor returns a new AdhocRequestExecutor.
func NewAdhocRequestExecutor(ctx context.Context, store Store, bus messaging.MessageBus) *AdhocRequestExecutor {
	ctx, cancel := context.WithCancel(ctx)
	executor := &AdhocRequestExecutor{
		adhocQueue: store.NewQueue(adhocQueueName),
		store:      store,
		bus:        bus,
		ctx:        ctx,
		cancel:     cancel,
	}
	go executor.listenQueue(ctx)
	return executor
}

// Stop calls the context cancel function to stop the AdhocRequestExecutor.
func (a *AdhocRequestExecutor) Stop() {
	a.cancel()
}

func (a *AdhocRequestExecutor) listenQueue(ctx context.Context) {
	for {
		select {
		case <-a.ctx.Done():
			return
		default:
		}
		// listen to the queue, unmarshal value into a check request, and execute it
		item, err := a.adhocQueue.Dequeue(ctx)
		if err != nil {
			a.listenQueueErr <- err
			continue
		}
		var check types.CheckConfig
		if err := json.Unmarshal([]byte(item.Value), &check); err != nil {
			a.listenQueueErr <- fmt.Errorf("unable to process invalid check: %s", err)
			if ackErr := item.Ack(ctx); ackErr != nil {
				a.listenQueueErr <- ackErr
			}
			continue
		}

		if err := a.processCheck(ctx, &check); err != nil {
			a.listenQueueErr <- err
			if nackErr := item.Nack(ctx); nackErr != nil {
				a.listenQueueErr <- nackErr
			}
			continue
		}
		if err = item.Ack(ctx); err != nil {
			a.listenQueueErr <- err
			continue
		}
	}
}

func (a *AdhocRequestExecutor) processCheck(ctx context.Context, check *types.CheckConfig) error {
	if check.ProxyRequests != nil {
		// get entities by namespace
		entities, err := a.store.GetEntities(ctx)
		if err != nil {
			return err
		}
		// call matchedEntities
		if matchedEntities := matchEntities(entities, check.ProxyRequests); len(matchedEntities) != 0 {
			if err := a.proxyCheck(matchedEntities, check); err != nil {
				logger.Error(err)
			} else {
				logger.Info("no matching entities, check will not be published")
			}
		}
	}
	return a.executeCheck(check)
}

func (a *AdhocRequestExecutor) proxyCheck(entities []*types.Entity, check *types.CheckConfig) error {
	var err error
	splay := float64(0)
	numEntities := float64(len(entities))
	if check.ProxyRequests.Splay {
		if splay, err = calculateSplayInterval(check, numEntities); err != nil {
			return err
		}
	}

	for _, entity := range entities {
		time.Sleep(time.Duration(splay) * time.Second)
		substitutedCheck, err := substituteEntityTokens(entity, check)
		if err != nil {
			return err
		}
		if err := a.executeCheck(substitutedCheck); err != nil {
			return err
		}
	}
	return nil
}

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
