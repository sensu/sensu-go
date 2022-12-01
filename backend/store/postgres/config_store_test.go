package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/selector"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/stretchr/testify/assert"
)

const (
	defaultNamespace = "default"
	entityName       = "__test-entity__"
)

func testWithPostgresConfigStore(t testing.TB, fn func(p storev2.ConfigStore)) {
	t.Helper()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		s := NewConfigStore(db)
		fn(s)
	})
}

func TestConfigStore_CreateOrUpdate(t *testing.T) {
	ec := &corev3.EntityConfig{}
	ec.GetTypeMeta()

	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		entity, err := getEntity(ctx, s, "default", entityName)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrNotFound{}, err)
		assert.Nil(t, entity)

		toCreate := corev3.FixtureEntityConfig(entityName)
		for i := 0; i < 4; i++ {
			toCreate.Metadata.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err = createOrUpdateEntity(ctx, s, toCreate)
		assert.NoError(t, err)

		entity, err = getEntity(ctx, s, defaultNamespace, entityName)
		assert.NoError(t, err)
		assert.Equal(t, entityName, entity.Metadata.Name)
		assert.Equal(t, 4, len(entity.Metadata.Labels))

		delete(toCreate.Metadata.Labels, "label-0")
		delete(toCreate.Metadata.Labels, "label-2")
		err = createOrUpdateEntity(ctx, s, toCreate)
		assert.NoError(t, err)

		entity, err = getEntity(ctx, s, "default", entityName)
		assert.NoError(t, err)
		assert.Equal(t, entityName, entity.Metadata.Name)
		assert.Equal(t, 2, len(entity.Metadata.Labels))
	})
}

func TestConfigStore_CreateIfNotExists(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		entity, err := getEntity(ctx, s, "default", entityName)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrNotFound{}, err)
		assert.Nil(t, entity)

		toCreate := corev3.FixtureEntityConfig(entityName)
		for i := 0; i < 4; i++ {
			toCreate.Metadata.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err = createIfNotExists(ctx, s, toCreate)
		assert.NoError(t, err)

		entity, err = getEntity(ctx, s, defaultNamespace, entityName)
		assert.NoError(t, err)
		assert.Equal(t, entityName, entity.Metadata.Name)
		assert.Equal(t, 4, len(entity.Metadata.Labels))

		delete(toCreate.Metadata.Labels, "label-0")
		delete(toCreate.Metadata.Labels, "label-2")
		err = createIfNotExists(ctx, s, toCreate)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrAlreadyExists{}, err)
	})
}

func TestConfigStore_UpdateIfExists(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		entity, err := getEntity(ctx, s, "default", entityName)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrNotFound{}, err)
		assert.Nil(t, entity)

		toCreate := corev3.FixtureEntityConfig(entityName)
		for i := 0; i < 4; i++ {
			toCreate.Metadata.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err = updateIfExists(ctx, s, toCreate)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrNotFound{}, err)

		err = createOrUpdateEntity(ctx, s, toCreate)
		assert.NoError(t, err)

		delete(toCreate.Metadata.Labels, "label-0")
		delete(toCreate.Metadata.Labels, "label-2")
		err = updateIfExists(ctx, s, toCreate)
		assert.NoError(t, err)

		entity, err = getEntity(ctx, s, defaultNamespace, entityName)
		assert.NoError(t, err)
		assert.Equal(t, entityName, entity.Metadata.Name)
		assert.Equal(t, 2, len(entity.Metadata.Labels))
	})
}

func TestConfigStore_Delete(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		toCreate := corev3.FixtureEntityConfig(entityName)
		for i := 0; i < 4; i++ {
			toCreate.Metadata.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err := deleteEntity(ctx, s, defaultNamespace, entityName)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrNotFound{}, err)

		err = createOrUpdateEntity(ctx, s, toCreate)
		assert.NoError(t, err)

		_, err = getEntity(ctx, s, defaultNamespace, entityName)
		assert.NoError(t, err)

		err = deleteEntity(ctx, s, defaultNamespace, entityName)
		assert.NoError(t, err)
	})
}

func TestConfigStore_Exists(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		toCreate := corev3.FixtureEntityConfig(entityName)
		for i := 0; i < 4; i++ {
			toCreate.Metadata.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		exists, err := entityExists(ctx, s, defaultNamespace, entityName)
		assert.NoError(t, err)
		assert.False(t, exists)

		err = createOrUpdateEntity(ctx, s, toCreate)
		assert.NoError(t, err)

		exists, err = entityExists(ctx, s, defaultNamespace, entityName)
		assert.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestConfigStore_Get(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		entity, err := getEntity(ctx, s, "default", entityName)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrNotFound{}, err)
		assert.Nil(t, entity)

		toCreate := corev3.FixtureEntityConfig(entityName)
		for i := 0; i < 4; i++ {
			toCreate.Metadata.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err = createIfNotExists(ctx, s, toCreate)
		assert.NoError(t, err)

		entity, err = getEntity(ctx, s, defaultNamespace, entityName)
		assert.NoError(t, err)
		assert.Equal(t, entityName, entity.Metadata.Name)
		assert.Equal(t, 4, len(entity.Metadata.Labels))
	})
}

func TestConfigStore_List(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		entities, err := listEntities(ctx, s, defaultNamespace, &store.SelectionPredicate{})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(entities))

		toCreate := corev3.FixtureEntityConfig(entityName)
		for i := 0; i < 4; i++ {
			toCreate.Metadata.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		for i := 0; i < 100; i++ {
			toCreate.Metadata.Name = fmt.Sprintf("%s%d__", entityName, i)
			err := createIfNotExists(ctx, s, toCreate)
			assert.NoError(t, err)
		}

		for i := 0; i < 10; i++ {
			toCreate.Metadata.Name = entityName
			toCreate.Metadata.Namespace = fmt.Sprintf("ns%d", i)
			err := createIfNotExists(ctx, s, toCreate)
			assert.NoError(t, err)
		}

		entities, err = listEntities(ctx, s, defaultNamespace, &store.SelectionPredicate{})
		assert.NoError(t, err)
		assert.Equal(t, 100, len(entities))

		entities, err = listEntities(ctx, s, defaultNamespace, &store.SelectionPredicate{Limit: 20})
		assert.NoError(t, err)
		assert.Equal(t, 20, len(entities))

		entities, err = listEntities(ctx, s, "", &store.SelectionPredicate{})
		assert.NoError(t, err)
		assert.Equal(t, 110, len(entities))
	})
}

func TestConfigStore_List_WithSelectors(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		for i := 0; i < 100; i++ {
			toCreate := corev3.FixtureEntityConfig(fmt.Sprintf("%s%d", entityName, i))
			toCreate.Metadata.Labels[fmt.Sprintf("label-mod-key-%d", i%3)] = "value"
			toCreate.Metadata.Labels["label-mod-value"] = fmt.Sprintf("value-%d", i%3)
			toCreate.Metadata.Labels["label-flat"] = fmt.Sprintf("value-%d", i)
			toCreate.Metadata.Labels["label-const"] = "const-value"
			toCreate.User = fmt.Sprintf("user-%d", (i+2)%3)

			err := createIfNotExists(context.Background(), s, toCreate)
			assert.NoError(t, err)
		}

		tests := []struct {
			name                string
			selektor            *selector.Selector
			expectError         bool
			expectedEntityCount int
			expectedEntityNames []string
		}{
			{
				name: "entity name label -in- selector",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"label-flat", selector.InOperator, []string{"value-6", "value-22"}, selector.OperationTypeLabelSelector},
						{"metadata.name", selector.InOperator, []string{entityName + "22", entityName + "45"}, selector.OperationTypeFieldSelector},
					},
				},
				expectError:         false,
				expectedEntityCount: 1,
				expectedEntityNames: []string{entityName + "22"},
			},
			{
				name: "entity name label -in- selector",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"label-flat", selector.InOperator, []string{"value-6", "value-22"}, selector.OperationTypeLabelSelector},
						{"metadata.name", selector.InOperator, []string{entityName + "22", entityName + "45"}, selector.OperationTypeFieldSelector},
					},
				},
				expectError:         false,
				expectedEntityCount: 1,
				expectedEntityNames: []string{entityName + "22"},
			},
			{
				name: "entity name label -in- selector",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"label-flat", selector.InOperator, []string{"value-6", "value-22"}, selector.OperationTypeLabelSelector},
					},
				},
				expectError:         false,
				expectedEntityCount: 2,
				expectedEntityNames: []string{entityName + "6", entityName + "22"},
			},
			{
				name: "entity name field -in- selector",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"metadata.name", selector.InOperator, []string{entityName + "6", entityName + "22"}, selector.OperationTypeFieldSelector},
					},
				},
				expectError:         false,
				expectedEntityCount: 2,
				expectedEntityNames: []string{entityName + "6", entityName + "22"},
			},
			{
				name: "entity name field -match- selector",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"metadata.name", selector.MatchesOperator, []string{fmt.Sprintf("%s%d", entityName, 65)}, selector.OperationTypeFieldSelector},
					},
				},
				expectError:         false,
				expectedEntityCount: 1,
				expectedEntityNames: []string{entityName + "65"},
			},
			{
				name: "label -match- selector",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"label-flat", selector.MatchesOperator, []string{"value-65"}, selector.OperationTypeLabelSelector},
					},
				},
				expectError:         false,
				expectedEntityCount: 1,
				expectedEntityNames: []string{entityName + "65"},
			},
			{
				name: "field and label -match- selectors",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"metadata.name", selector.MatchesOperator, []string{entityName + "6"}, selector.OperationTypeFieldSelector},
						{"label-mod-key-0", selector.MatchesOperator, []string{"value"}, selector.OperationTypeLabelSelector},
					},
				},
				expectError:         false,
				expectedEntityCount: 5,
				expectedEntityNames: []string{entityName + "6", entityName + "60", entityName + "63", entityName + "66", entityName + "69"},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				ctx := context.Background()
				selCtx := selector.ContextWithSelector(ctx, test.selektor)
				entities, err := listEntities(selCtx, s, defaultNamespace, &store.SelectionPredicate{})
				if test.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				entityCount, err := countEntities(selCtx, s, defaultNamespace)
				if test.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expectedEntityCount, len(entities))
				assert.Equal(t, test.expectedEntityCount, entityCount)
				for _, name := range test.expectedEntityNames {
					var found bool
					for _, entity := range entities {
						if entity.Metadata.Name == name {
							found = true
							break
						}
					}
					assert.True(t, found, fmt.Sprintf("entity not found: %s", name))
				}
			})
		}
	})
}

func TestConfigStore_Count(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		entities, err := listEntities(ctx, s, defaultNamespace, &store.SelectionPredicate{})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(entities))

		toCreate := corev3.FixtureEntityConfig(entityName)
		for i := 0; i < 4; i++ {
			toCreate.Metadata.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		for i := 0; i < 100; i++ {
			toCreate.Metadata.Name = fmt.Sprintf("%s%d__", entityName, i)
			err := createIfNotExists(ctx, s, toCreate)
			assert.NoError(t, err)
		}

		for i := 0; i < 10; i++ {
			toCreate.Metadata.Name = entityName
			toCreate.Metadata.Namespace = fmt.Sprintf("ns%d", i)
			err := createIfNotExists(ctx, s, toCreate)
			assert.NoError(t, err)
		}

		ct, err := countEntities(ctx, s, defaultNamespace)
		assert.NoError(t, err)
		assert.Equal(t, 100, ct)

		ct, err = countEntities(ctx, s, "")
		assert.NoError(t, err)
		assert.Equal(t, 110, ct)
	})
}

func TestConfigStore_Patch(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		toCreate := corev3.FixtureEntityConfig(entityName)
		for i := 0; i < 4; i++ {
			toCreate.Metadata.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}
	})
}

func TestConfigStore_Watch(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		stor, ok := s.(*ConfigStore)
		if !ok {
			t.Error("expected config store")
			return
		}

		stor.watchInterval = time.Millisecond * 10
		stor.watchTxnWindow = time.Second

		ctx := context.Background()
		entity := corev3.FixtureEntityConfig("my-entity")
		watchReq := storev2.ResourceRequest{
			APIVersion: entity.GetTypeMeta().APIVersion,
			Type:       entity.GetTypeMeta().Type,
		}
		watchChannel := s.Watch(ctx, watchReq)
		select {
		case record, ok := <-watchChannel:
			t.Errorf("expected watch channel to be empty. Got %v, %v", record, ok)
		default:
			// OK
		}

		// create notification
		err := createOrUpdateEntity(ctx, s, entity)
		if err != nil {
			t.Error(err)
			return
		}
		select {
		case watchEvents, ok := <-watchChannel:
			if !ok {
				t.Error("watcher closed unexpectedly")
				return
			}
			if len(watchEvents) != 1 {
				t.Error("expected 1 watch event")
				return
			}
			assert.Equal(t, storev2.WatchCreate, watchEvents[0].Type)

		case <-time.After(5 * time.Second):
			t.Fatalf("no watch event received before timeout")
		}

		// update notification
		entity.Metadata.Labels["new-label"] = "new-value"
		err = createOrUpdateEntity(ctx, s, entity)
		if err != nil {
			t.Error(err)
			return
		}
		select {
		case watchEvents, ok := <-watchChannel:
			if !ok {
				t.Error("watcher closed unexpectedly")
				return
			}
			if len(watchEvents) != 1 {
				t.Error("expected 1 watch event")
				return
			}
			assert.Equal(t, storev2.WatchUpdate, watchEvents[0].Type)

		case <-time.After(5 * time.Second):
			t.Fatalf("no watch event received before timeout")
		}

		// delete notification
		err = deleteEntity(ctx, s, entity.Metadata.Namespace, entity.Metadata.Name)
		if err != nil {
			t.Error(err)
			return
		}
		if err != nil {
			t.Error(err)
			return
		}
		select {
		case watchEvents, ok := <-watchChannel:
			if !ok {
				t.Error("watcher closed unexpectedly")
				return
			}
			if len(watchEvents) != 1 {
				t.Error("expected 1 watch event")
				return
			}
			assert.Equal(t, storev2.WatchDelete, watchEvents[0].Type)

		case <-time.After(5 * time.Second):
			t.Fatalf("no watch event received before timeout")
		}
	})
}

func createOrUpdateEntity(ctx context.Context, pgStore storev2.ConfigStore, entity *corev3.EntityConfig) error {
	req := storev2.ResourceRequest{
		APIVersion: entity.GetTypeMeta().APIVersion,
		Type:       entity.GetTypeMeta().Type,
		Namespace:  entity.Metadata.Namespace,
		Name:       entity.Metadata.Name,
		StoreName:  "entity_configs",
		SortOrder:  0,
	}

	wrapper, err := wrapEntity(entity)
	if err != nil {
		return err
	}

	return pgStore.CreateOrUpdate(ctx, req, wrapper)
}

func createIfNotExists(ctx context.Context, pgStore storev2.ConfigStore, entity *corev3.EntityConfig) error {
	req := storev2.ResourceRequest{
		APIVersion: entity.GetTypeMeta().APIVersion,
		Type:       entity.GetTypeMeta().Type,
		Namespace:  entity.Metadata.Namespace,
		Name:       entity.Metadata.Name,
		StoreName:  "entity_configs",
		SortOrder:  0,
	}

	wrapper, err := wrapEntity(entity)
	if err != nil {
		return err
	}

	return pgStore.CreateIfNotExists(ctx, req, wrapper)
}

func countEntities(ctx context.Context, pgStore storev2.ConfigStore, namespace string) (int, error) {
	entityConfig := corev3.EntityConfig{}
	typeMeta := entityConfig.GetTypeMeta()
	req := storev2.ResourceRequest{
		Namespace:  namespace,
		StoreName:  "entity_configs",
		APIVersion: typeMeta.APIVersion,
		Type:       typeMeta.Type,
	}

	return pgStore.Count(ctx, req)
}

func listEntities(ctx context.Context, pgStore storev2.ConfigStore, namespace string, predicate *store.SelectionPredicate) ([]*corev3.EntityConfig, error) {
	entityConfig := corev3.EntityConfig{}
	typeMeta := entityConfig.GetTypeMeta()
	req := storev2.ResourceRequest{
		Namespace:  namespace,
		Name:       "",
		StoreName:  "entity_configs",
		APIVersion: typeMeta.APIVersion,
		Type:       typeMeta.Type,
		SortOrder:  0,
	}

	list, err := pgStore.List(ctx, req, predicate)
	if err != nil {
		return nil, err
	}

	res, err := list.Unwrap()
	if err != nil {
		return nil, err
	}

	entities := make([]*corev3.EntityConfig, 0, len(res))
	for _, ent := range res {
		entity, ok := ent.(*corev3.EntityConfig)
		if !ok {
			return nil, errors.New("not an entity config")
		}
		entities = append(entities, entity)
	}

	return entities, nil
}

func getEntity(ctx context.Context, pgStore storev2.ConfigStore, namespace, name string) (*corev3.EntityConfig, error) {
	entityConfig := corev3.EntityConfig{}
	typeMeta := entityConfig.GetTypeMeta()
	req := storev2.ResourceRequest{
		Namespace:  namespace,
		Name:       name,
		StoreName:  "entity_configs",
		APIVersion: typeMeta.APIVersion,
		Type:       typeMeta.Type,
		SortOrder:  0,
	}

	entityWrapper, err := pgStore.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	res, err := entityWrapper.Unwrap()
	if err != nil {
		return nil, err
	}

	entity, ok := res.(*corev3.EntityConfig)
	if !ok {
		return nil, errors.New("resource is not an entity")
	}

	return entity, nil
}

func deleteEntity(ctx context.Context, pgStore storev2.ConfigStore, namespace, name string) error {
	entityConfig := corev3.EntityConfig{}
	typeMeta := entityConfig.GetTypeMeta()
	req := storev2.ResourceRequest{
		Namespace:  namespace,
		Name:       name,
		StoreName:  "entity_configs",
		APIVersion: typeMeta.APIVersion,
		Type:       typeMeta.Type,
		SortOrder:  0,
	}

	return pgStore.Delete(ctx, req)
}

func entityExists(ctx context.Context, pgStore storev2.ConfigStore, namespace, name string) (bool, error) {
	entityConfig := corev3.EntityConfig{}
	typeMeta := entityConfig.GetTypeMeta()
	req := storev2.ResourceRequest{
		Namespace:  namespace,
		Name:       name,
		StoreName:  "entity_configs",
		APIVersion: typeMeta.APIVersion,
		Type:       typeMeta.Type,
		SortOrder:  0,
	}

	return pgStore.Exists(ctx, req)
}

func updateIfExists(ctx context.Context, pgStore storev2.ConfigStore, entityConfig *corev3.EntityConfig) error {
	req := storev2.ResourceRequest{
		APIVersion: entityConfig.GetTypeMeta().APIVersion,
		Type:       entityConfig.GetTypeMeta().Type,
		Namespace:  entityConfig.Metadata.Namespace,
		Name:       entityConfig.Metadata.Name,
		StoreName:  "entity_configs",
		SortOrder:  0,
	}

	wrapper, err := wrapEntity(entityConfig)
	if err != nil {
		return err
	}

	return pgStore.UpdateIfExists(ctx, req, wrapper)
}

func wrapEntity(entity *corev3.EntityConfig) (*wrap.Wrapper, error) {
	jsonEntity, err := json.Marshal(entity)
	if err != nil {
		return nil, err
	}
	typeMeta := entity.GetTypeMeta()

	return &wrap.Wrapper{
		TypeMeta:    &typeMeta,
		Encoding:    wrap.Encoding_json,
		Compression: wrap.Compression_none,
		Value:       jsonEntity,
	}, nil
}
