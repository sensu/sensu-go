package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
)

var (
	entityStateStoreName        = new(corev3.EntityState).StoreName()
	entityStateUniqueConstraint = "entity_state_unique"
)

type namespacedEntityStates map[string][]*corev3.EntityState

type uniqueEntityStates map[uniqueResource]*corev3.EntityState

type EntityStateStore struct {
	db *pgxpool.Pool
}

func NewEntityStateStore(db *pgxpool.Pool) *EntityStateStore {
	return &EntityStateStore{
		db: db,
	}
}

// Create creates an entity state using the given entity state struct.
func (s *EntityStateStore) CreateIfNotExists(ctx context.Context, state *corev3.EntityState) error {
	if err := state.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	namespace := state.Metadata.Namespace
	name := state.Metadata.Name
	wrapper := WrapEntityState(state).(*EntityStateWrapper)
	params := wrapper.SQLParams()

	if _, err := s.db.Exec(ctx, createIfNotExistsEntityStateQuery, params...); err != nil {
		pgError, ok := err.(*pgconn.PgError)
		if ok {
			switch pgError.ConstraintName {
			case entityStateUniqueConstraint:
				return &store.ErrAlreadyExists{
					Key: entityStateStoreKey(namespace, name),
				}
			}
		}
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

// CreateOrUpdate creates an entity state or updates it if it already exists.
func (s *EntityStateStore) CreateOrUpdate(ctx context.Context, state *corev3.EntityState) error {
	if err := state.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	wrapper := WrapEntityState(state).(*EntityStateWrapper)
	params := wrapper.SQLParams()

	if _, err := s.db.Exec(ctx, createOrUpdateEntityStateQuery, params...); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

// Delete soft deletes an entity state using the given namespace & name.
func (s *EntityStateStore) Delete(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		return &store.ErrNotValid{Err: errors.New("must specify namespace")}
	}
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	result, err := s.db.Exec(ctx, deleteEntityStateQuery, namespace, name)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	affected := result.RowsAffected()
	if affected < 1 {
		return &store.ErrNotFound{Key: entityStateStoreKey(namespace, name)}
	}
	return nil
}

// Exists determines if an entity state exists.
func (s *EntityStateStore) Exists(ctx context.Context, namespace, name string) (bool, error) {
	if namespace == "" {
		return false, &store.ErrNotValid{Err: errors.New("must specify namespace")}
	}
	if name == "" {
		return false, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	row := s.db.QueryRow(ctx, existsEntityStateQuery, namespace, name)
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

// Get retrieves an entity state by a given namespace & name.
func (s *EntityStateStore) Get(ctx context.Context, namespace, name string) (*corev3.EntityState, error) {
	if namespace == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify namespace")}
	}
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	row := s.db.QueryRow(ctx, getEntityStateQuery, namespace, name)
	wrapper := &EntityStateWrapper{}
	if err := row.Scan(wrapper.SQLParams()...); err != nil {
		if err == pgx.ErrNoRows {
			return nil, &store.ErrNotFound{Key: entityStateStoreKey(namespace, name)}
		}
		return nil, &store.ErrInternal{Message: err.Error()}
	}

	var state corev3.EntityState
	if err := wrapper.unwrapIntoEntityState(&state); err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}

	return &state, nil
}

// GetMultiple retrieves multiple entity states for a given mapping of namespace
// and names.
func (s *EntityStateStore) GetMultiple(ctx context.Context, resources namespacedResourceNames) (uniqueEntityStates, error) {
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

	states := uniqueEntityStates{}
	for namespace, resourceNames := range resources {
		rows, err := tx.Query(ctx, getEntityStatesQuery, namespace, resourceNames)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			wrapper := &EntityStateWrapper{}
			if err := rows.Scan(wrapper.SQLParams()...); err != nil {
				if err == pgx.ErrNoRows {
					continue
				}
				return nil, &store.ErrInternal{Message: err.Error()}
			}

			state := &corev3.EntityState{}
			if err := wrapper.unwrapIntoEntityState(state); err != nil {
				return nil, &store.ErrInternal{Message: err.Error()}
			}

			key := uniqueResource{
				Name:      state.Metadata.Name,
				Namespace: namespace,
			}
			states[key] = state
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction for GetMultiple()")
	}

	return states, nil
}

// HardDelete hard deletes an entity state using the given namespace & name.
func (s *EntityStateStore) HardDelete(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		return &store.ErrNotValid{Err: errors.New("must specify namespace")}
	}
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	result, err := s.db.Exec(ctx, hardDeleteEntityStateQuery, namespace, name)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	affected := result.RowsAffected()
	if affected < 1 {
		return &store.ErrNotFound{Key: entityStateStoreKey(namespace, name)}
	}
	return nil
}

// HardDeleted determines if an entity state has been hard deleted.
func (s *EntityStateStore) HardDeleted(ctx context.Context, namespace, name string) (bool, error) {
	if namespace == "" {
		return false, &store.ErrNotValid{Err: errors.New("must specify namespace")}
	}
	if name == "" {
		return false, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	row := s.db.QueryRow(ctx, hardDeletedEntityStateQuery, namespace, name)
	var deleted bool
	if err := row.Scan(&deleted); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, &store.ErrInternal{Message: err.Error()}
	}
	return deleted, nil
}

// List returns all entity states. A nil slice with no error is returned if
// none were found.
func (s *EntityStateStore) List(ctx context.Context, namespace string, pred *store.SelectionPredicate) ([]*corev3.EntityState, error) {
	query := listEntityStateQuery
	if pred != nil && pred.Descending {
		query = listEntityStateDescQuery
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

	rows, rerr := s.db.Query(ctx, query, sqlNamespace, limit, offset)
	if rerr != nil {
		return nil, &store.ErrInternal{Message: rerr.Error()}
	}
	defer rows.Close()

	var states []*corev3.EntityState
	var rowCount int64
	for rows.Next() {
		rowCount++
		wrapper := &EntityStateWrapper{}
		if err := rows.Scan(wrapper.SQLParams()...); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		if err := rows.Err(); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		state := &corev3.EntityState{}
		if err := wrapper.unwrapIntoEntityState(state); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		states = append(states, state)
	}
	if pred != nil {
		if rowCount < pred.Limit {
			pred.Continue = ""
		}
	}

	return states, nil
}

func (s *EntityStateStore) Patch(ctx context.Context, namespace, name string, patcher patch.Patcher, conditions *store.ETagCondition) (fErr error) {
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

	wrapper := &EntityStateWrapper{}
	row := tx.QueryRow(ctx, getEntityStateQuery, namespace, name)
	if err := row.Scan(wrapper.SQLParams()...); err != nil {
		if err == pgx.ErrNoRows {
			return &store.ErrNotFound{Key: entityStateStoreKey(namespace, name)}
		}
		return &store.ErrInternal{Message: err.Error()}
	}

	state := corev3.EntityState{}
	if err := wrapper.unwrapIntoEntityState(&state); err != nil {
		return &store.ErrDecode{Key: entityStateStoreKey(namespace, name), Err: err}
	}

	// Now determine the etag for the stored resource
	etag, err := store.ETag(state)
	if err != nil {
		return err
	}

	if conditions != nil {
		if !store.CheckIfMatch(conditions.IfMatch, etag) {
			return &store.ErrPreconditionFailed{Key: entityStateStoreKey(namespace, name)}
		}
		if !store.CheckIfNoneMatch(conditions.IfNoneMatch, etag) {
			return &store.ErrPreconditionFailed{Key: entityStateStoreKey(namespace, name)}
		}
	}

	// Encode the stored resource to the JSON format
	originalJSON, err := json.Marshal(state)
	if err != nil {
		return err
	}

	// Apply the patch to our original document (stored resource)
	patchedResource, err := patcher.Patch(originalJSON)
	if err != nil {
		return err
	}

	// Decode the resulting JSON document back into our resource
	if err := json.Unmarshal(patchedResource, &state); err != nil {
		return err
	}

	// Validate the resource
	if err := state.Validate(); err != nil {
		return err
	}

	// Patch the resource
	patchedWrapper := WrapEntityState(&state).(*EntityStateWrapper)
	params := patchedWrapper.SQLParams()

	if _, err := tx.Exec(ctx, createOrUpdateEntityStateQuery, params...); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

// UpdateIfExists updates a given entity state.
func (s *EntityStateStore) UpdateIfExists(ctx context.Context, state *corev3.EntityState) error {
	namespace := state.Metadata.Namespace
	name := state.Metadata.Name
	wrapper := WrapEntityState(state).(*EntityStateWrapper)
	params := wrapper.SQLParams()

	rows, err := s.db.Query(ctx, updateIfExistsEntityStateQuery, params...)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	defer rows.Close()

	rowCount := 0
	for rows.Next() {
		rowCount++
	}
	if rowCount == 0 {
		return &store.ErrNotFound{Key: entityStateStoreKey(namespace, name)}
	}

	return nil
}

func entityStateStoreKey(namespace, name string) string {
	return fmt.Sprintf("%s(namespace, name)=(%s, %s)", entityStateStoreName, namespace, name)
}
