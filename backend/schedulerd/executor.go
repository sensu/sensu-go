package schedulerd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	time "github.com/echlebek/timeproxy"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/types"
	"github.com/sirupsen/logrus"
)

var (
	adhocQueueName = "adhocRequest"
)

// Executor executes scheduled or adhoc checks
type Executor interface {
	processCheck(ctx context.Context, check *corev2.CheckConfig) error
	getEntities(ctx context.Context) ([]cache.Value, error)
	publishProxyCheckRequests(entities []*corev2.Entity, check *corev2.CheckConfig) error
	execute(check *corev2.CheckConfig) error
	buildRequest(check *corev2.CheckConfig) (*corev2.CheckRequest, error)
}

// CheckExecutor executes scheduled checks in the check scheduler
type CheckExecutor struct {
	bus                    messaging.MessageBus
	store                  store.Store
	namespace              string
	entityCache            *cache.Resource
	secretsProviderManager *secrets.ProviderManager
}

// NewCheckExecutor creates a new check executor
func NewCheckExecutor(bus messaging.MessageBus, namespace string, store store.Store, cache *cache.Resource, secretsProviderManager *secrets.ProviderManager) *CheckExecutor {
	return &CheckExecutor{bus: bus, namespace: namespace, store: store, entityCache: cache, secretsProviderManager: secretsProviderManager}
}

// ProcessCheck processes a check by publishing its proxy requests (if any)
// and publishing the check itself
func (c *CheckExecutor) processCheck(ctx context.Context, check *corev2.CheckConfig) error {
	return processCheck(ctx, c, check)
}

func (c *CheckExecutor) getEntities(ctx context.Context) ([]cache.Value, error) {
	return c.entityCache.Get(store.NewNamespaceFromContext(ctx)), nil
}

func (c *CheckExecutor) publishProxyCheckRequests(entities []*corev2.Entity, check *corev2.CheckConfig) error {
	return publishProxyCheckRequests(c, entities, check)
}

func (c *CheckExecutor) execute(check *corev2.CheckConfig) error {
	// Ensure the check is configured to publish check requests
	if !check.Publish {
		return nil
	}

	var err error
	request, err := c.buildRequest(check)
	if err != nil {
		return err
	}

	for _, sub := range check.Subscriptions {
		topic := messaging.SubscriptionTopic(check.Namespace, sub)
		logger.WithFields(logrus.Fields{
			"check": check.Name,
			"topic": topic,
		}).Debug("sending check request")

		if pubErr := c.bus.Publish(topic, request); pubErr != nil {
			logger.WithError(pubErr).Error("error publishing check request")
			err = pubErr
		}
	}

	return err
}

func (c *CheckExecutor) executeOnEntity(check *corev2.CheckConfig, entity string) error {
	// Ensure the check is configured to publish check requests
	if !check.Publish {
		return nil
	}

	var err error
	request, err := c.buildRequest(check)
	if err != nil {
		return err
	}

	topic := messaging.SubscriptionTopic(check.Namespace, fmt.Sprintf("entity:%s", entity))
	logger.WithFields(logrus.Fields{
		"check": check.Name,
		"topic": topic,
	}).Debug("sending check request")

	return c.bus.Publish(topic, request)
}

func (c *CheckExecutor) buildRequest(check *corev2.CheckConfig) (*corev2.CheckRequest, error) {
	return buildRequest(check, c.store, c.secretsProviderManager)
}

func assetIsRelevant(asset *corev2.Asset, assets []string) bool {
	for _, assetName := range assets {
		if strings.HasPrefix(asset.Name, assetName) {
			return true
		}
	}

	return false
}

func hookIsRelevant(hook *corev2.HookConfig, check *corev2.CheckConfig) bool {
	for _, checkHook := range check.CheckHooks {
		for _, hookName := range checkHook.Hooks {
			if hookName == hook.Name {
				return true
			}
		}
	}

	return false
}

// AdhocRequestExecutor takes new check requests from the adhoc queue and runs
// them
type AdhocRequestExecutor struct {
	adhocQueue             types.Queue
	store                  store.Store
	bus                    messaging.MessageBus
	ctx                    context.Context
	cancel                 context.CancelFunc
	listenQueueErr         chan error
	entityCache            *cache.Resource
	secretsProviderManager *secrets.ProviderManager
}

// NewAdhocRequestExecutor returns a new AdhocRequestExecutor.
func NewAdhocRequestExecutor(ctx context.Context, store store.Store, queue types.Queue, bus messaging.MessageBus, cache *cache.Resource, secretsProviderManager *secrets.ProviderManager) *AdhocRequestExecutor {
	ctx, cancel := context.WithCancel(ctx)
	executor := &AdhocRequestExecutor{
		adhocQueue:             queue,
		store:                  store,
		bus:                    bus,
		ctx:                    ctx,
		cancel:                 cancel,
		entityCache:            cache,
		secretsProviderManager: secretsProviderManager,
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
		if err := a.ctx.Err(); err != nil {
			return
		}
		// listen to the queue, unmarshal value into a check request, and execute it
		item, err := a.adhocQueue.Dequeue(ctx)
		if err != nil {
			a.listenQueueErr <- err
			continue
		}
		var check corev2.CheckConfig
		if err := json.NewDecoder(strings.NewReader(item.Value())).Decode(&check); err != nil {
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
func (a *AdhocRequestExecutor) processCheck(ctx context.Context, check *corev2.CheckConfig) error {
	return processCheck(ctx, a, check)
}

func (a *AdhocRequestExecutor) getEntities(ctx context.Context) ([]cache.Value, error) {
	return a.entityCache.Get(store.NewNamespaceFromContext(ctx)), nil
}

func (a *AdhocRequestExecutor) publishProxyCheckRequests(entities []*corev2.Entity, check *corev2.CheckConfig) error {
	return publishProxyCheckRequests(a, entities, check)
}

func (a *AdhocRequestExecutor) execute(check *corev2.CheckConfig) error {
	var err error
	request, err := a.buildRequest(check)
	if err != nil {
		return err
	}

	for _, sub := range check.Subscriptions {
		topic := messaging.SubscriptionTopic(check.Namespace, sub)
		logger.WithFields(logrus.Fields{
			"check": check.Name,
			"topic": topic,
		}).Debug("sending check request")

		if pubErr := a.bus.Publish(topic, request); pubErr != nil {
			logger.WithError(pubErr).Error("error publishing check request")
			err = pubErr
		}
	}
	return err
}

func (a *AdhocRequestExecutor) buildRequest(check *corev2.CheckConfig) (*corev2.CheckRequest, error) {
	return buildRequest(check, a.store, a.secretsProviderManager)
}

func publishProxyCheckRequests(e Executor, entities []*corev2.Entity, check *corev2.CheckConfig) error {
	var splay time.Duration
	if check.ProxyRequests.Splay {
		var err error
		if splay, err = calculateSplayInterval(check, len(entities)); err != nil {
			return err
		}
	}

	for _, entity := range entities {
		time.Sleep(splay)
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

func processCheck(ctx context.Context, executor Executor, check *corev2.CheckConfig) error {
	fields := logrus.Fields{
		"check":     check.Name,
		"namespace": check.Namespace,
	}
	if check.ProxyRequests != nil {
		// get entities by namespace
		entities, err := executor.getEntities(ctx)
		if err != nil {
			return err
		}
		// publish proxy requests on matching entities
		if matchedEntities := matchEntities(entities, check.ProxyRequests); len(matchedEntities) != 0 {
			if err := executor.publishProxyCheckRequests(matchedEntities, check); err != nil {
				logger.WithFields(fields).WithError(err).Error("error publishing proxy check requests")
			}
		} else {
			logger.WithFields(fields).Warn("no matching entities, check will not be published")
		}
	} else {
		return executor.execute(check)
	}
	return nil
}

func processRoundRobinCheck(ctx context.Context, executor *CheckExecutor, check *corev2.CheckConfig, proxyEntities []*corev2.Entity, agentEntities []string) error {
	if check.ProxyRequests != nil {
		return publishRoundRobinProxyCheckRequests(executor, check, proxyEntities, agentEntities)
	}
	for _, entity := range agentEntities {
		if err := executor.executeOnEntity(check, entity); err != nil {
			return err
		}
	}
	return nil
}

func publishRoundRobinProxyCheckRequests(executor *CheckExecutor, check *corev2.CheckConfig, proxyEntities []*corev2.Entity, agentEntities []string) error {
	var splay time.Duration
	if check.ProxyRequests.Splay {
		var err error
		if splay, err = calculateSplayInterval(check, len(proxyEntities)); err != nil {
			return err
		}
	}

	for i, proxyEntity := range proxyEntities {
		now := time.Now()
		agentEntity := agentEntities[i]
		substitutedCheck, err := substituteProxyEntityTokens(proxyEntity, check)
		if err != nil {
			return err
		}
		if err := executor.executeOnEntity(substitutedCheck, agentEntity); err != nil {
			return err
		}
		dreamtime := splay - time.Now().Sub(now)
		time.Sleep(dreamtime)
	}
	return nil
}

func buildRequest(check *corev2.CheckConfig, s store.Store, secretsProviderManager *secrets.ProviderManager) (*corev2.CheckRequest, error) {
	request := &corev2.CheckRequest{}
	request.Config = check
	request.HookAssets = make(map[string]*corev2.AssetList)

	// Prepare log entry
	fields := logrus.Fields{
		"namespace": check.Namespace,
		"mutator":   check.Name,
		"assets":    check.RuntimeAssets,
	}

	secrets, err := secretsProviderManager.SubSecrets(check.Command)
	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to retrieve secrets for check")
		return nil, err
	}
	request.Secrets = secrets

	ctx := corev2.SetContextFromResource(context.Background(), check)

	assets, err := s.GetAssets(ctx, &store.SelectionPredicate{})
	if err != nil {
		return nil, err
	}

	// Guard against iterating over assets if there are no assets associated with
	// the check in the first place.
	if len(check.RuntimeAssets) != 0 {
		// Filter out assets that are irrelevant
		for _, asset := range assets {
			if assetIsRelevant(asset, check.RuntimeAssets) {
				request.Assets = append(request.Assets, *asset)
			}
		}
	}

	// Guard against iterating over hooks if there are no hooks associated with
	// the check in the first place.
	if len(check.CheckHooks) != 0 {
		// Explode hooks; get hooks & filter out those that are irrelevant
		hooks, err := s.GetHookConfigs(ctx, &store.SelectionPredicate{})
		if err != nil {
			return nil, err
		}

		for _, hook := range hooks {
			if hookIsRelevant(hook, check) {
				request.Hooks = append(request.Hooks, *hook)
				if len(hook.RuntimeAssets) != 0 {
					assetList := &corev2.AssetList{}
					for _, asset := range assets {
						if assetIsRelevant(asset, hook.RuntimeAssets) {
							assetList.Assets = append(assetList.Assets, *asset)
						}
					}
					request.HookAssets[hook.Name] = assetList
				}
			}
		}
	}

	request.Issued = time.Now().Unix()

	return request, nil
}
