package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/poll"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

var (
	entityConfigStoreName        = new(corev3.EntityConfig).StoreName()
	entityConfigUniqueConstraint = "entity_config_unique"
)

type uniqueEntityConfigs map[uniqueResource]*corev3.EntityConfig

type EntityConfigStore struct {
	db DBI
}

func NewEntityConfigStore(db DBI) *EntityConfigStore {
	return &EntityConfigStore{
		db: db,
	}
}

type namespacedResourceNames map[string][]string

const beginningOfTime = "0001-01-01T00:00:00Z"

// CreateIfNotExists creates an entity config using the given entity config struct.
func (s *EntityConfigStore) CreateIfNotExists(ctx context.Context, config *corev3.EntityConfig) error {
	if err := config.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	namespace := config.Metadata.Namespace
	name := config.Metadata.Name
	wrapper := WrapEntityConfig(config).(*EntityConfigWrapper)
	params := wrapper.SQLParams()

	if _, err := s.db.Exec(ctx, createIfNotExistsEntityConfigQuery, params...); err != nil {
		pgError, ok := err.(*pgconn.PgError)
		if ok {
			switch pgError.ConstraintName {
			case entityConfigUniqueConstraint:
				return &store.ErrAlreadyExists{
					Key: entityConfigStoreKey(namespace, name),
				}
			}
		}
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

// CreateOrUpdate creates an entity config or updates it if it already exists.
func (s *EntityConfigStore) CreateOrUpdate(ctx context.Context, config *corev3.EntityConfig) error {
	if err := config.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	wrapper := WrapEntityConfig(config).(*EntityConfigWrapper)
	params := wrapper.SQLParams()

	if _, err := s.db.Exec(ctx, createOrUpdateEntityConfigQuery, params...); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

// Delete soft deletes an entity config using the given namespace & name.
func (s *EntityConfigStore) Delete(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		return &store.ErrNotValid{Err: errors.New("must specify namespace")}
	}
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	result, err := s.db.Exec(ctx, deleteEntityConfigQuery, namespace, name)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	affected := result.RowsAffected()
	if affected < 1 {
		return &store.ErrNotFound{Key: entityConfigStoreKey(namespace, name)}
	}
	return nil
}

// Exists determines if an entity config exists.
func (s *EntityConfigStore) Exists(ctx context.Context, namespace, name string) (bool, error) {
	if namespace == "" {
		return false, &store.ErrNotValid{Err: errors.New("must specify namespace")}
	}
	if name == "" {
		return false, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	row := s.db.QueryRow(ctx, existsEntityConfigQuery, namespace, name)
	var found bool
	err := row.Scan(&found)
	if err == nil {
		return found, nil
	}
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return false, &store.ErrInternal{Message: err.Error()}
}

// Get retrieves an entity config by a given namespace & name.
func (s *EntityConfigStore) Get(ctx context.Context, namespace, name string) (*corev3.EntityConfig, error) {
	if namespace == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify namespace")}
	}
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	row := s.db.QueryRow(ctx, getEntityConfigQuery, namespace, name)
	wrapper := &EntityConfigWrapper{}
	if err := row.Scan(wrapper.SQLParams()...); err != nil {
		if err == pgx.ErrNoRows {
			return nil, &store.ErrNotFound{Key: entityConfigStoreKey(namespace, name)}
		}
		return nil, &store.ErrInternal{Message: err.Error()}
	}

	var config corev3.EntityConfig
	if err := wrapper.unwrapIntoEntityConfig(&config); err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}

	return &config, nil
}

// GetMultiple retrieves multiple entity configs for a given mapping of namespace
// and names.
func (s *EntityConfigStore) GetMultiple(ctx context.Context, resources namespacedResourceNames) (uniqueEntityConfigs, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Commit(ctx); err != nil && err != pgx.ErrTxClosed {
			logger.WithError(err).Error("error committing transaction for GetMultiple()")
		}
	}()

	configs := uniqueEntityConfigs{}
	for namespace, resourceNames := range resources {
		rows, err := tx.Query(ctx, getEntityConfigsQuery, namespace, resourceNames)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			wrapper := &EntityConfigWrapper{}
			if err := rows.Scan(wrapper.SQLParams()...); err != nil {
				if err == pgx.ErrNoRows {
					continue
				}
				return nil, &store.ErrInternal{Message: err.Error()}
			}

			config := &corev3.EntityConfig{}
			if err := wrapper.unwrapIntoEntityConfig(config); err != nil {
				return nil, &store.ErrInternal{Message: err.Error()}
			}

			key := uniqueResource{
				Name:      config.Metadata.Name,
				Namespace: namespace,
			}
			configs[key] = config
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction for GetMultiple()")
	}

	return configs, nil
}

// HardDelete hard deletes an entity config using the given namespace & name.
func (s *EntityConfigStore) HardDelete(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		return &store.ErrNotValid{Err: errors.New("must specify namespace")}
	}
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	result, err := s.db.Exec(ctx, hardDeleteEntityConfigQuery, namespace, name)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	affected := result.RowsAffected()
	if affected < 1 {
		return &store.ErrNotFound{Key: entityConfigStoreKey(namespace, name)}
	}
	return nil
}

// HardDeleted determines if an entity config has been hard deleted.
func (s *EntityConfigStore) HardDeleted(ctx context.Context, namespace, name string) (bool, error) {
	if namespace == "" {
		return false, &store.ErrNotValid{Err: errors.New("must specify namespace")}
	}
	if name == "" {
		return false, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	row := s.db.QueryRow(ctx, hardDeletedEntityConfigQuery, namespace, name)
	var deleted bool
	if err := row.Scan(&deleted); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, &store.ErrInternal{Message: err.Error()}
	}
	return deleted, nil
}

// List returns all entity configs. A nil slice with no error is returned if
// none were found.
func (s *EntityConfigStore) List(ctx context.Context, namespace string, pred *store.SelectionPredicate) ([]*corev3.EntityConfig, error) {
	if pred == nil {
		pred = &store.SelectionPredicate{}
	}
	query := listEntityConfigQuery
	if pred.Descending {
		query = listEntityConfigDescQuery
	}

	if pred.UpdatedSince == "" {
		pred.UpdatedSince = beginningOfTime
	}

	var updatedSince time.Time
	if err := updatedSince.UnmarshalText([]byte(pred.UpdatedSince)); err != nil {
		return nil, &store.ErrNotValid{Err: fmt.Errorf("bad UpdatedSince time: %s", err)}
	}

	limit, offset, err := getLimitAndOffset(pred)
	if err != nil {
		return nil, err
	}

	var sqlNamespace sql.NullString
	if namespace != "" {
		sqlNamespace.String = namespace
		sqlNamespace.Valid = true
	}

	rows, rerr := s.db.Query(ctx, query, sqlNamespace, limit, offset, pred.IncludeDeletes, updatedSince)
	if rerr != nil {
		return nil, &store.ErrInternal{Message: rerr.Error()}
	}
	defer rows.Close()

	var configs []*corev3.EntityConfig
	var rowCount int64
	for rows.Next() {
		rowCount++
		wrapper := &EntityConfigWrapper{}
		if err := rows.Scan(wrapper.SQLParams()...); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		if err := rows.Err(); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		config := &corev3.EntityConfig{}
		if err := wrapper.unwrapIntoEntityConfig(config); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		configs = append(configs, config)
	}
	if pred != nil {
		if rowCount < pred.Limit {
			pred.Continue = ""
		}
	}

	return configs, nil
}

// Count returns the count of EntityConfigs found matching the namespace
// and entity class provided.
func (s *EntityConfigStore) Count(ctx context.Context, namespace, entityClass string) (int, error) {
	var sqlNamespace sql.NullString
	if namespace != "" {
		sqlNamespace.String = namespace
		sqlNamespace.Valid = true
	}
	var sqlEntityClass sql.NullString
	if entityClass != "" {
		sqlEntityClass.String = entityClass
		sqlEntityClass.Valid = true
	}

	row := s.db.QueryRow(ctx, countEntityConfigQuery, sqlNamespace, sqlEntityClass)
	var ct int
	if err := row.Scan(&ct); err != nil {
		return 0, err
	}
	return ct, nil
}

func (s *EntityConfigStore) Patch(ctx context.Context, namespace, name string, patcher patch.Patcher) (fErr error) {
	if namespace == "" {
		return &store.ErrNotValid{Err: errors.New("must specify namespace")}
	}
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	tx, txerr := s.db.Begin(ctx)
	if txerr != nil {
		return &store.ErrInternal{Message: txerr.Error()}
	}
	defer func() {
		if fErr == nil {
			fErr = tx.Commit(ctx)
			return
		}
		if txerr := tx.Rollback(ctx); txerr != nil {
			fErr = txerr
		}
	}()

	wrapper := &EntityConfigWrapper{}
	row := tx.QueryRow(ctx, getEntityConfigQuery, namespace, name)
	if err := row.Scan(wrapper.SQLParams()...); err != nil {
		if err == pgx.ErrNoRows {
			return &store.ErrNotFound{Key: entityConfigStoreKey(namespace, name)}
		}
		return &store.ErrInternal{Message: err.Error()}
	}

	config := corev3.EntityConfig{}
	if err := wrapper.unwrapIntoEntityConfig(&config); err != nil {
		return &store.ErrDecode{Key: entityConfigStoreKey(namespace, name), Err: err}
	}

	// Now determine the etag for the stored resource
	etag := storev2.ETagFromStruct(config)

	ifMatch := storev2.IfMatchFromContext(ctx)
	ifNoneMatch := storev2.IfNoneMatchFromContext(ctx)

	if ifMatch != nil {
		if !ifMatch.Matches(etag) {
			return &store.ErrPreconditionFailed{Key: entityConfigStoreKey(namespace, name)}
		}
	}

	if !ifNoneMatch.Matches(etag) {
		return &store.ErrPreconditionFailed{Key: entityConfigStoreKey(namespace, name)}
	}

	// Encode the stored resource to the JSON format
	originalJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}

	// Apply the patch to our original document (stored resource)
	patchedResource, err := patcher.Patch(originalJSON)
	if err != nil {
		return err
	}

	// Decode the resulting JSON document back into our resource
	if err := json.Unmarshal(patchedResource, &config); err != nil {
		return err
	}

	// Validate the resource
	if err := config.Validate(); err != nil {
		return err
	}

	// Patch the resource
	patchedWrapper := WrapEntityConfig(&config).(*EntityConfigWrapper)
	params := patchedWrapper.SQLParams()

	if _, err := tx.Exec(ctx, createOrUpdateEntityConfigQuery, params...); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

// UpdateIfExists updates a given entity config.
func (s *EntityConfigStore) UpdateIfExists(ctx context.Context, config *corev3.EntityConfig) error {
	namespace := config.Metadata.Namespace
	name := config.Metadata.Name
	wrapper := WrapEntityConfig(config).(*EntityConfigWrapper)
	params := wrapper.SQLParams()

	rows, err := s.db.Query(ctx, updateIfExistsEntityConfigQuery, params...)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	defer rows.Close()

	rowCount := 0
	for rows.Next() {
		rowCount++
	}
	if rowCount == 0 {
		return &store.ErrNotFound{Key: entityConfigStoreKey(namespace, name)}
	}

	return nil
}

func entityConfigStoreKey(namespace, name string) string {
	return fmt.Sprintf("%s(namespace, name)=(%s, %s)", entityConfigStoreName, namespace, name)
}

func (s *EntityConfigStore) Watch(ctx context.Context, namespace, name string) <-chan []storev2.WatchEvent {
	req := storev2.ResourceRequest{
		Namespace:  namespace,
		Name:       name,
		APIVersion: "core/v3",
		Type:       "EntityConfig",
		StoreName:  "entity_configs",
	}
	return NewWatcher(s, 0, 0).Watch(ctx, req)
}

func (s *EntityConfigStore) GetPoller(req storev2.ResourceRequest) (poll.Table, error) {
	return newEntityConfigPoller(req, s.db)
}
