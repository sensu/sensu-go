package schedulerd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/queue"
	"github.com/sensu/sensu-go/types"
)

var (
	adhocQueueName = "adhocRequest"
)

// Executor executes scheduled or adhoc checks
type Executor interface {
	processCheck(ctx context.Context, check *types.CheckConfig) error
	getEntities(ctx context.Context) ([]*types.Entity, error)
	publishProxyCheckRequests(entities []*types.Entity, check *types.CheckConfig) error
	execute(check *types.CheckConfig) error
	buildRequest(check *types.CheckConfig) *types.CheckRequest
	setState(state *SchedulerState)
}

// CheckExecutor executes scheduled checks in the check scheduler
type CheckExecutor struct {
	bus          messaging.MessageBus
	state        *SchedulerState
	roundRobin   *roundRobinScheduler
	organization string
	environment  string
}

// NewCheckExecutor creates a new check executor
func NewCheckExecutor(bus messaging.MessageBus, roundRobin *roundRobinScheduler, org string, env string) *CheckExecutor {
	return &CheckExecutor{bus: bus, roundRobin: roundRobin, organization: org, environment: env}
}

// ProcessCheck processes a check by publishing its proxy requests (if any)
// and publishing the check itself
func (c *CheckExecutor) processCheck(ctx context.Context, check *types.CheckConfig) error {
	return processCheck(ctx, c, check)
}

func (c *CheckExecutor) getEntities(ctx context.Context) ([]*types.Entity, error) {
	return c.state.GetEntitiesInNamespace(c.organization, c.environment), nil
}

func (c *CheckExecutor) publishProxyCheckRequests(entities []*types.Entity, check *types.CheckConfig) error {
	return publishProxyCheckRequests(c, entities, check)
}

func (c *CheckExecutor) execute(check *types.CheckConfig) error {
	// Ensure the check is configured to publish check requests
	if !check.Publish {
		return nil
	}

	var err error
	request := c.buildRequest(check)

	for _, sub := range check.Subscriptions {
		org, env := check.Organization, check.Environment
		topic := messaging.SubscriptionTopic(org, env, sub)
		if check.RoundRobin {
			msg := &roundRobinMessage{
				subscription: topic,
				req:          request,
			}
			_, err = c.roundRobin.Schedule(msg)
			if err != nil {
				logger.WithError(err).Error("error scheduling round robin request")
			}
			continue
		}
		logger.Debugf("sending check request for %s on topic %s", check.Name, topic)

		if pubErr := c.bus.Publish(topic, request); pubErr != nil {
			logger.WithError(pubErr).Error("error publishing check request")
			err = pubErr
		}
	}

	return err
}

func (c *CheckExecutor) buildRequest(check *types.CheckConfig) *types.CheckRequest {
	request := &types.CheckRequest{}
	request.Config = check

	// Guard against iterating over assets if there are no assets associated with
	// the check in the first place.
	if len(check.RuntimeAssets) != 0 {
		// Explode assets; get assets & filter out those that are irrelevant
		allAssets := c.state.GetAssetsInNamespace(check.Organization)
		for _, asset := range allAssets {
			if assetIsRelevant(asset, check) {
				request.Assets = append(request.Assets, *asset)
			}
		}
	}

	// Guard against iterating over hooks if there are no hooks associated with
	// the check in the first place.
	if len(check.CheckHooks) != 0 {
		// Explode hooks; get hooks & filter out those that are irrelevant
		allHooks := c.state.GetHooksInNamespace(check.Organization, check.Environment)
		for _, hook := range allHooks {
			if hookIsRelevant(hook, check) {
				request.Hooks = append(request.Hooks, *hook)
			}
		}
	}

	return request
}

func assetIsRelevant(asset *types.Asset, check *types.CheckConfig) bool {
	for _, assetName := range check.RuntimeAssets {
		if strings.HasPrefix(asset.Name, assetName) {
			return true
		}
	}

	return false
}

func hookIsRelevant(hook *types.HookConfig, check *types.CheckConfig) bool {
	for _, checkHook := range check.CheckHooks {
		for _, hookName := range checkHook.Hooks {
			if hookName == hook.Name {
				return true
			}
		}
	}

	return false
}

func (c *CheckExecutor) setState(state *SchedulerState) {
	c.state = state
}

// AdhocRequestExecutor takes new check requests from the adhoc queue and runs
// them
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
		if err = json.Unmarshal([]byte(item.Value), &check); err != nil {
			a.listenQueueErr <- fmt.Errorf("unable to process invalid check: %s", err)
			if ackErr := item.Ack(ctx); ackErr != nil {
				a.listenQueueErr <- ackErr
			}
			continue
		}

		if err = a.processCheck(ctx, &check); err != nil {
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

// processCheck processes a check by publishing its proxy requests (if any)
// and publishing the check itself
func (a *AdhocRequestExecutor) processCheck(ctx context.Context, check *types.CheckConfig) error {
	return processCheck(ctx, a, check)
}

func (a *AdhocRequestExecutor) getEntities(ctx context.Context) ([]*types.Entity, error) {
	return a.store.GetEntities(ctx)
}

func (a *AdhocRequestExecutor) publishProxyCheckRequests(entities []*types.Entity, check *types.CheckConfig) error {
	return publishProxyCheckRequests(a, entities, check)
}

func (a *AdhocRequestExecutor) execute(check *types.CheckConfig) error {
	request := a.buildRequest(check)
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

func (a *AdhocRequestExecutor) buildRequest(check *types.CheckConfig) *types.CheckRequest {
	return &types.CheckRequest{}
}

func (a *AdhocRequestExecutor) setState(state *SchedulerState) {}

func publishProxyCheckRequests(e Executor, entities []*types.Entity, check *types.CheckConfig) error {
	var err error
	splay := float64(0)
	numEntities := float64(len(entities))
	if check.ProxyRequests.Splay {
		if splay, err = calculateSplayInterval(check, numEntities); err != nil {
			return err
		}
	}

	for _, entity := range entities {
		time.Sleep(time.Duration(time.Millisecond * time.Duration(splay*1000)))
		substitutedCheck, err := substituteProxyEntityTokens(entity, check)
		if err != nil {
			return err
		}
		if err := e.execute(substitutedCheck); err != nil {
			return err
		}
	}
	return nil
}

func processCheck(ctx context.Context, executor Executor, check *types.CheckConfig) error {
	if check.ProxyRequests != nil {
		// get entities by namespace
		entities, err := executor.getEntities(ctx)
		if err != nil {
			return err
		}
		// publish proxy requests on matching entities
		if matchedEntities := matchEntities(entities, check.ProxyRequests); len(matchedEntities) != 0 {
			if err := executor.publishProxyCheckRequests(matchedEntities, check); err != nil {
				logger.Error(err)
			}
		} else {
			logger.Info("no matching entities, check will not be published")
		}
	} else {
		return executor.execute(check)
	}
	return nil
}
