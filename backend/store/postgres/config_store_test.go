package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	v2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultNamespace = "default"
	entityName       = "__test-entity__"
)

func testWithPostgresConfigStore(t *testing.T, fn func(p storev2.Interface)) {

	_ = os.Setenv("PG_URL", "user=4b6f0bad-b9e0-42f2-a62b-de1c087f426c password=70b1be6d-3c84-48d2-bcc7-d6ca45af4530 host=localhost port=5432 dbname=sensudb sslmode=disable")

	t.Helper()
	if testing.Short() {
		t.Skip("skipping postgres test")
		return
	}
	pgURL := os.Getenv("PG_URL")
	if pgURL == "" {
		t.Skip("skipping postgres test")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, err := pgxpool.Connect(ctx, pgURL)
	require.NoError(t, err)

	dbName := "sensuconfigdb" + strings.ReplaceAll(uuid.New().String(), "-", "")
	_, err = db.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", dbName))
	require.NoError(t, err)

	defer dropAll(context.Background(), dbName, pgURL)
	db.Close()

	testURL := fmt.Sprintf("%s dbname=%s ", pgURL, dbName)
	pgxConfig, err := pgxpool.ParseConfig(testURL)
	require.NoError(t, err)

	configDB, err := OpenConfigDB(ctx, pgxConfig, false)
	require.NoError(t, err)
	defer configDB.Close()

	s := NewConfigStore(configDB)
	fn(s)
}

func TestConfigStore_CreateOrUpdate(t *testing.T) {
	ec := &corev3.EntityConfig{}
	ec.GetTypeMeta()

	testWithPostgresConfigStore(t, func(s storev2.Interface) {
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
	testWithPostgresConfigStore(t, func(s storev2.Interface) {
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
	testWithPostgresConfigStore(t, func(s storev2.Interface) {
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
	testWithPostgresConfigStore(t, func(s storev2.Interface) {
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
	testWithPostgresConfigStore(t, func(s storev2.Interface) {
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
	testWithPostgresConfigStore(t, func(s storev2.Interface) {
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
	testWithPostgresConfigStore(t, func(s storev2.Interface) {
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

		entities, err = listEntities(ctx, s, defaultNamespace, &store.SelectionPredicate{})
		assert.NoError(t, err)
		assert.Equal(t, 100, len(entities))

		entities, err = listEntities(ctx, s, defaultNamespace, &store.SelectionPredicate{Limit: 20})
		assert.NoError(t, err)
		assert.Equal(t, 20, len(entities))
	})
}

func TestConfigStore_Patch(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.Interface) {
		toCreate := corev3.FixtureEntityConfig(entityName)
		for i := 0; i < 4; i++ {
			toCreate.Metadata.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		//err = createIfNotExists(ctx, s, toCreate)
		//assert.NoError(t, err)

	})
}

func createOrUpdateEntity(ctx context.Context, pgStore storev2.Interface, entity *corev3.EntityConfig) error {
	req := storev2.ResourceRequest{
		Namespace:   entity.Metadata.Namespace,
		Name:        entity.Metadata.Name,
		StoreName:   "entity_configs",
		Context:     ctx,
		SortOrder:   0,
		UsePostgres: true,
	}

	wrapper, err := wrapEntity(entity)
	if err != nil {
		return err
	}

	return pgStore.CreateOrUpdate(req, wrapper)
}

func createIfNotExists(ctx context.Context, pgStore storev2.Interface, entity *corev3.EntityConfig) error {
	req := storev2.ResourceRequest{
		Namespace:   entity.Metadata.Namespace,
		Name:        entity.Metadata.Name,
		StoreName:   "entity_configs",
		Context:     ctx,
		SortOrder:   0,
		UsePostgres: true,
	}

	wrapper, err := wrapEntity(entity)
	if err != nil {
		return err
	}

	return pgStore.CreateIfNotExists(req, wrapper)
}

func listEntities(ctx context.Context, pgStore storev2.Interface, namespace string, predicate *store.SelectionPredicate) ([]*corev3.EntityConfig, error) {
	req := storev2.ResourceRequest{
		Namespace:   namespace,
		Name:        "",
		StoreName:   "entity_configs",
		TypeMeta:    &v2.TypeMeta{APIVersion: "core/v3", Type: "EntityConfig"},
		Context:     ctx,
		SortOrder:   0,
		UsePostgres: true,
	}

	list, err := pgStore.List(req, predicate)
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

func getEntity(ctx context.Context, pgStore storev2.Interface, namespace, name string) (*corev3.EntityConfig, error) {
	req := storev2.ResourceRequest{
		Namespace:   namespace,
		Name:        name,
		StoreName:   "entity_configs",
		TypeMeta:    &v2.TypeMeta{APIVersion: "core/v3", Type: "EntityConfig"},
		Context:     ctx,
		SortOrder:   0,
		UsePostgres: true,
	}

	entityWrapper, err := pgStore.Get(req)
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

func deleteEntity(ctx context.Context, pgStore storev2.Interface, namespace, name string) error {
	req := storev2.ResourceRequest{
		Namespace:   namespace,
		Name:        name,
		StoreName:   "entity_configs",
		TypeMeta:    &v2.TypeMeta{APIVersion: "core/v3", Type: "EntityConfig"},
		Context:     ctx,
		SortOrder:   0,
		UsePostgres: true,
	}

	return pgStore.Delete(req)
}

func entityExists(ctx context.Context, pgStore storev2.Interface, namespace, name string) (bool, error) {
	req := storev2.ResourceRequest{
		Namespace:   namespace,
		Name:        name,
		StoreName:   "entity_configs",
		TypeMeta:    &v2.TypeMeta{APIVersion: "core/v3", Type: "EntityConfig"},
		Context:     ctx,
		SortOrder:   0,
		UsePostgres: true,
	}

	return pgStore.Exists(req)
}

func updateIfExists(ctx context.Context, pgStore storev2.Interface, entityConfig *corev3.EntityConfig) error {
	req := storev2.ResourceRequest{
		Namespace:   entityConfig.Metadata.Namespace,
		Name:        entityConfig.Metadata.Name,
		StoreName:   "entity_configs",
		Context:     ctx,
		SortOrder:   0,
		UsePostgres: true,
	}

	wrapper, err := wrapEntity(entityConfig)
	if err != nil {
		return err
	}

	return pgStore.UpdateIfExists(req, wrapper)
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
