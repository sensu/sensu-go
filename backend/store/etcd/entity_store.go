package etcd

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd/kvc"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	entityPathPrefix       = "entities"
	entityConfigPathPrefix = "entity_configs"
)

var (
	entityKeyBuilder       = store.NewKeyBuilder(entityPathPrefix)
	entityConfigKeyBuilder = store.NewKeyBuilder(entityConfigPathPrefix)
)

type entityContinueToken struct {
	ConfigContinue []byte `json:",omitempty"`
	StateContinue  []byte `json:",omitempty"`
}

func getEntityPath(entity *corev2.Entity) string {
	return entityKeyBuilder.WithResource(entity).Build(entity.Name)
}

// GetEntitiesPath gets the path of the entity store
func GetEntitiesPath(ctx context.Context, name string) string {
	return entityKeyBuilder.WithContext(ctx).Build(name)
}

// GetEntityConfigsPath gets the path of entity_configs in the store
func GetEntityConfigsPath(ctx context.Context, name string) string {
	return entityConfigKeyBuilder.WithContext(ctx).Build(name)
}

// DeleteEntity deletes an Entity.
func (s *Store) DeleteEntity(ctx context.Context, e *corev2.Entity) error {
	if err := e.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}
	state := &corev3.EntityState{
		Metadata: &e.ObjectMeta,
	}
	config := &corev3.EntityConfig{
		Metadata: &e.ObjectMeta,
	}
	stateReq := storev2.NewResourceRequestFromResource(ctx, state)
	configReq := storev2.NewResourceRequestFromResource(ctx, config)
	stateKey := etcdstore.StoreKey(stateReq)
	configKey := etcdstore.StoreKey(configReq)

	comparator := kvc.Comparisons()
	ops := []clientv3.Op{
		clientv3.OpDelete(stateKey),
		clientv3.OpDelete(configKey),
	}

	return kvc.Txn(ctx, s.client, comparator, ops...)
}

// DeleteEntityByName deletes an Entity by its name.
func (s *Store) DeleteEntityByName(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	state := &corev3.EntityState{
		Metadata: &corev2.ObjectMeta{
			Name:      name,
			Namespace: corev2.ContextNamespace(ctx),
		},
	}
	config := &corev3.EntityConfig{
		Metadata: &corev2.ObjectMeta{
			Name:      name,
			Namespace: corev2.ContextNamespace(ctx),
		},
	}
	stateReq := storev2.NewResourceRequestFromResource(ctx, state)
	configReq := storev2.NewResourceRequestFromResource(ctx, config)
	stateKey := etcdstore.StoreKey(stateReq)
	configKey := etcdstore.StoreKey(configReq)

	comparator := kvc.Comparisons(
		kvc.KeyIsFound(stateKey),
		kvc.KeyIsFound(configKey),
	)
	ops := []clientv3.Op{
		clientv3.OpDelete(stateKey),
		clientv3.OpDelete(configKey),
	}

	return kvc.Txn(ctx, s.client, comparator, ops...)
}

// GetEntityByName gets an Entity by its name.
func (s *Store) GetEntityByName(ctx context.Context, name string) (*corev2.Entity, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}
	state := &corev3.EntityState{
		Metadata: &corev2.ObjectMeta{
			Name:      name,
			Namespace: corev2.ContextNamespace(ctx),
		},
	}
	config := &corev3.EntityConfig{
		Metadata: &corev2.ObjectMeta{
			Name:      name,
			Namespace: corev2.ContextNamespace(ctx),
		},
	}
	stateReq := storev2.NewResourceRequestFromResource(ctx, state)
	configReq := storev2.NewResourceRequestFromResource(ctx, config)
	stateKey := etcdstore.StoreKey(stateReq)
	configKey := etcdstore.StoreKey(configReq)
	ops := []clientv3.Op{
		clientv3.OpGet(stateKey, clientv3.WithLimit(1)),
		clientv3.OpGet(configKey, clientv3.WithLimit(1)),
	}
	var resp *clientv3.TxnResponse
	err := kvc.Backoff(ctx).Retry(func(n int) (done bool, err error) {
		resp, err = s.client.Txn(ctx).Then(ops...).Commit()
		return kvc.RetryRequest(n, err)
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Responses) != 2 {
		panic("txn response with 2 ops did not return 2 responses")
	}
	stateResp := resp.Responses[0].GetResponseRange()
	if len(stateResp.Kvs) == 0 {
		return nil, nil
	}
	configResp := resp.Responses[1].GetResponseRange()
	if len(configResp.Kvs) == 0 {
		return nil, nil
	}
	var (
		configWrapper, stateWrapper wrap.Wrapper
		entityConfig                corev3.EntityConfig
		entityState                 corev3.EntityState
	)
	if err := proto.Unmarshal(stateResp.Kvs[0].Value, &stateWrapper); err != nil {
		return nil, &store.ErrDecode{Err: err, Key: stateKey}
	}
	if err := proto.Unmarshal(configResp.Kvs[0].Value, &configWrapper); err != nil {
		return nil, &store.ErrDecode{Err: err, Key: configKey}
	}
	if err := stateWrapper.UnwrapInto(&entityState); err != nil {
		return nil, &store.ErrDecode{Err: err, Key: stateKey}
	}
	if err := configWrapper.UnwrapInto(&entityConfig); err != nil {
		return nil, &store.ErrDecode{Err: err, Key: configKey}
	}
	entity, err := corev3.V3EntityToV2(&entityConfig, &entityState)
	if err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}
	return entity, nil
}

// GetEntities returns the entities for the namespace in the supplied context.
func (s *Store) GetEntities(ctx context.Context, pred *store.SelectionPredicate) ([]*corev2.Entity, error) {
	v2store := etcdstore.NewStore(s.client)
	namespace := corev2.ContextNamespace(ctx)
	stateReq := storev2.ResourceRequest{
		Namespace: namespace,
		Context:   ctx,
		StoreName: new(corev3.EntityState).StoreName(),
	}
	configReq := storev2.ResourceRequest{
		Namespace: namespace,
		Context:   ctx,
		StoreName: new(corev3.EntityConfig).StoreName(),
	}
	statePred := new(store.SelectionPredicate)
	configPred := new(store.SelectionPredicate)
	if pred != nil {
		if pred.Limit != 0 {
			statePred.Limit = pred.Limit
			configPred.Limit = pred.Limit
		}
		if pred.Continue != "" {
			var token entityContinueToken
			if err := json.Unmarshal([]byte(pred.Continue), &token); err != nil {
				return nil, &store.ErrNotValid{Err: err}
			}
			statePred.Continue = string(token.StateContinue)
			configPred.Continue = string(token.ConfigContinue)
		}
		if pred.Ordering == corev2.EntitySortName {
			dir := storev2.SortAscend
			if pred.Descending {
				dir = storev2.SortDescend
			}
			stateReq.SortOrder = dir
			configReq.SortOrder = dir
		}
	}
	stateList, err := v2store.List(stateReq, statePred)
	if err != nil {
		return nil, err
	}
	configList, err := v2store.List(configReq, configPred)
	if err != nil {
		return nil, err
	}
	if pred != nil {
		var token entityContinueToken
		token.ConfigContinue = []byte(configPred.Continue)
		token.StateContinue = []byte(statePred.Continue)
		b, _ := json.Marshal(token)
		cont := string(b)
		if cont == "{}" {
			cont = ""
		}
		pred.Continue = cont
	}
	states := make([]corev3.EntityState, stateList.Len())
	if err := stateList.UnwrapInto(&states); err != nil {
		return nil, &store.ErrDecode{Err: err, Key: etcdstore.StoreKey(stateReq)}
	}
	configs := make([]corev3.EntityConfig, configList.Len())
	if err := configList.UnwrapInto(&configs); err != nil {
		return nil, &store.ErrDecode{Err: err, Key: etcdstore.StoreKey(configReq)}
	}
	return entitiesFromConfigAndState(configs, states)
}

func entitiesFromConfigAndState(configs []corev3.EntityConfig, states []corev3.EntityState) ([]*corev2.Entity, error) {
	result := make([]*corev2.Entity, 0, len(states))
	var i, j int
	for i < len(states) && j < len(configs) {
		switch states[i].Metadata.Cmp(configs[j].Metadata) {
		case corev2.MetaLess:
			// there is a state without a corresponding config
			i++
		case corev2.MetaEqual:
			state := &states[i]
			config := &configs[j]
			entity, err := corev3.V3EntityToV2(config, state)
			if err != nil {
				return nil, &store.ErrNotValid{Err: err}
			}
			result = append(result, entity)
			i++
			j++
		case corev2.MetaGreater:
			// there is a config without a corresponding state, create anyway
			result = append(result, entityFromConfigOnly(&configs[j]))
			j++
		}
	}
	for j < len(configs) {
		result = append(result, entityFromConfigOnly(&configs[j]))
		j++
	}
	return result, nil
}

func entityFromConfigOnly(config *corev3.EntityConfig) *corev2.Entity {
	state := corev3.NewEntityState(config.Metadata.Namespace, config.Metadata.Name)
	entity, _ := corev3.V3EntityToV2(config, state)
	return entity
}

// UpdateEntity updates an Entity.
func (s *Store) UpdateEntity(ctx context.Context, e *corev2.Entity) error {
	namespace := e.Namespace
	if namespace == "" {
		namespace = corev2.ContextNamespace(ctx)
	}
	cfg, state := corev3.V2EntityToV3(e)
	cfg.Metadata.Namespace = namespace
	state.Metadata.Namespace = namespace
	stateReq := storev2.NewResourceRequestFromResource(ctx, state)
	configReq := storev2.NewResourceRequestFromResource(ctx, cfg)
	stateKey := etcdstore.StoreKey(stateReq)
	configKey := etcdstore.StoreKey(configReq)
	wrappedState, err := wrap.Resource(state)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}
	wrappedConfig, err := wrap.Resource(cfg)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}
	stateMsg, err := proto.Marshal(wrappedState)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}
	configMsg, err := proto.Marshal(wrappedConfig)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}

	comparator := kvc.Comparisons(
		kvc.NamespaceExists(namespace),
	)
	ops := []clientv3.Op{
		clientv3.OpPut(configKey, string(configMsg)),
		clientv3.OpPut(stateKey, string(stateMsg)),
	}

	return kvc.Txn(ctx, s.client, comparator, ops...)
}
