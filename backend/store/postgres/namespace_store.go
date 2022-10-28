package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
)

var (
	namespaceUniqueConstraint = "namespace_unique"
)

type uniqueNamespaces map[uniqueResource]*corev3.Namespace

type NamespaceStore struct {
	db DBI
}

func NewNamespaceStore(db DBI) *NamespaceStore {
	return &NamespaceStore{
		db: db,
	}
}

// Create creates a namespace using the given namespace struct.
func (s *NamespaceStore) CreateIfNotExists(ctx context.Context, namespace *corev3.Namespace) error {
	if err := namespace.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	wrapper := WrapNamespace(namespace).(*NamespaceWrapper)
	params := wrapper.SQLParams()

	if _, err := s.db.Exec(ctx, createIfNotExistsNamespaceQuery, params...); err != nil {
		pgError, ok := err.(*pgconn.PgError)
		if ok {
			switch pgError.ConstraintName {
			case namespaceUniqueConstraint:
				return &store.ErrAlreadyExists{
					Key: namespaceStoreKey(namespace.Metadata.Name),
				}
			}
		}
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

// CreateOrUpdate creates a namespace or updates it if it already exists.
func (s *NamespaceStore) CreateOrUpdate(ctx context.Context, namespace *corev3.Namespace) error {
	if err := namespace.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	wrapper := WrapNamespace(namespace).(*NamespaceWrapper)
	params := wrapper.SQLParams()

	if _, err := s.db.Exec(ctx, createOrUpdateNamespaceQuery, params...); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

// Delete soft deletes a namespace using the given namespace name.
func (s *NamespaceStore) Delete(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	empty, err := s.IsEmpty(ctx, name)
	if err != nil {
		return err
	}
	if !empty {
		return &store.ErrNamespaceNotEmpty{Namespace: name}
	}

	result, err := s.db.Exec(ctx, deleteNamespaceQuery, name)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	affected := result.RowsAffected()
	if affected < 1 {
		return &store.ErrNotFound{Key: namespaceStoreKey(name)}
	}
	return nil
}

// Exists determines if a namespace exists.
func (s *NamespaceStore) Exists(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	row := s.db.QueryRow(ctx, existsNamespaceQuery, name)
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

// Get retrieves a namespace by a given name.
func (s *NamespaceStore) Get(ctx context.Context, name string) (*corev3.Namespace, error) {
	if name == "" {
		return nil, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	row := s.db.QueryRow(ctx, getNamespaceQuery, name)
	wrapper := &NamespaceWrapper{}
	if err := row.Scan(wrapper.SQLParams()...); err != nil {
		if err == pgx.ErrNoRows {
			return nil, &store.ErrNotFound{Key: namespaceStoreKey(name)}
		}
		return nil, &store.ErrInternal{Message: err.Error()}
	}

	var namespace corev3.Namespace
	if err := wrapper.unwrapIntoNamespace(&namespace); err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}

	return &namespace, nil
}

// GetMultiple retrieves multiple namespaces for a given list of names.
func (s *NamespaceStore) GetMultiple(ctx context.Context, resources []string) (uniqueNamespaces, error) {
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

	namespaces := uniqueNamespaces{}
	rows, err := tx.Query(ctx, getNamespacesQuery, resources)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		wrapper := &NamespaceWrapper{}
		if err := rows.Scan(wrapper.SQLParams()...); err != nil {
			if err == pgx.ErrNoRows {
				continue
			}
			return nil, &store.ErrInternal{Message: err.Error()}
		}

		namespace := &corev3.Namespace{}
		if err := wrapper.unwrapIntoNamespace(namespace); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}

		key := uniqueResource{
			Name:      namespace.Metadata.Name,
			Namespace: "",
		}
		namespaces[key] = namespace
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction for GetMultiple()")
	}

	return namespaces, nil
}

// HardDelete hard deletes a namespace using the given namespace name.
func (s *NamespaceStore) HardDelete(ctx context.Context, name string) error {
	if name == "" {
		return &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	empty, err := s.IsEmpty(ctx, name)
	if err != nil {
		return err
	}
	if !empty {
		return &store.ErrNamespaceNotEmpty{Namespace: name}
	}

	result, err := s.db.Exec(ctx, hardDeleteNamespaceQuery, name)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	affected := result.RowsAffected()
	if affected < 1 {
		return &store.ErrNotFound{Key: namespaceStoreKey(name)}
	}
	return nil
}

// HardDeleted determines if a namespace has been hard deleted.
func (s *NamespaceStore) HardDeleted(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	row := s.db.QueryRow(ctx, hardDeletedNamespaceQuery, name)
	var deleted bool
	if err := row.Scan(&deleted); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, &store.ErrInternal{Message: err.Error()}
	}
	return deleted, nil
}

// List returns all namespaces. A nil slice with no error is returned if none
// were found.
func (s *NamespaceStore) List(ctx context.Context, pred *store.SelectionPredicate) ([]*corev3.Namespace, error) {
	query := listNamespaceQuery
	if pred != nil && pred.Descending {
		query = listNamespaceDescQuery
	}

	limit, offset, err := getLimitAndOffset(pred)
	if err != nil {
		return nil, err
	}

	rows, rerr := s.db.Query(ctx, query, limit, offset)
	if rerr != nil {
		return nil, &store.ErrInternal{Message: rerr.Error()}
	}
	defer rows.Close()

	namespaces := []*corev3.Namespace{}
	var rowCount int64
	for rows.Next() {
		rowCount++
		wrapper := &NamespaceWrapper{}
		if err := rows.Scan(wrapper.SQLParams()...); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		if err := rows.Err(); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		namespace := &corev3.Namespace{}
		if err := wrapper.unwrapIntoNamespace(namespace); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		namespaces = append(namespaces, namespace)
	}
	if pred != nil {
		if rowCount < pred.Limit {
			pred.Continue = ""
		}
	}

	return namespaces, nil
}

func (s *NamespaceStore) Count(ctx context.Context) (int, error) {
	row := s.db.QueryRow(ctx, countNamespaceQuery)

	var ct int
	err := row.Scan(&ct)
	return ct, err
}

func (s *NamespaceStore) Patch(ctx context.Context, name string, patcher patch.Patcher, conditions *store.ETagCondition) (fErr error) {
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

	wrapper := &NamespaceWrapper{}
	row := tx.QueryRow(ctx, getNamespaceQuery, name)
	if err := row.Scan(wrapper.SQLParams()...); err != nil {
		if err == pgx.ErrNoRows {
			return &store.ErrNotFound{Key: namespaceStoreKey(name)}
		}
		return &store.ErrInternal{Message: err.Error()}
	}

	namespace := corev3.Namespace{}
	if err := wrapper.unwrapIntoNamespace(&namespace); err != nil {
		return &store.ErrDecode{Key: namespaceStoreKey(name), Err: err}
	}

	// Now determine the etag for the stored resource
	etag, err := store.ETag(namespace)
	if err != nil {
		return err
	}

	if conditions != nil {
		if !store.CheckIfMatch(conditions.IfMatch, etag) {
			return &store.ErrPreconditionFailed{Key: namespaceStoreKey(name)}
		}
		if !store.CheckIfNoneMatch(conditions.IfNoneMatch, etag) {
			return &store.ErrPreconditionFailed{Key: namespaceStoreKey(name)}
		}
	}

	// Encode the stored resource to the JSON format
	originalJSON, err := json.Marshal(namespace)
	if err != nil {
		return err
	}

	// Apply the patch to our original document (stored resource)
	patchedResource, err := patcher.Patch(originalJSON)
	if err != nil {
		return err
	}

	// Decode the resulting JSON document back into our resource
	if err := json.Unmarshal(patchedResource, &namespace); err != nil {
		return err
	}

	// Validate the resource
	if err := namespace.Validate(); err != nil {
		return err
	}

	// Patch the resource
	patchedWrapper := WrapNamespace(&namespace).(*NamespaceWrapper)
	params := patchedWrapper.SQLParams()

	if _, err := tx.Exec(ctx, createOrUpdateNamespaceQuery, params...); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

// UpdateIfExists updates a given namespace.
func (s *NamespaceStore) UpdateIfExists(ctx context.Context, namespace *corev3.Namespace) error {
	wrapper := WrapNamespace(namespace).(*NamespaceWrapper)
	params := wrapper.SQLParams()

	rows, err := s.db.Query(ctx, updateIfExistsNamespaceQuery, params...)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	defer rows.Close()

	rowCount := 0
	for rows.Next() {
		rowCount++
	}
	if rowCount == 0 {
		return &store.ErrNotFound{Key: namespaceStoreKey(namespace.Metadata.Name)}
	}

	return nil
}

func (s *NamespaceStore) IsEmpty(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, &store.ErrNotValid{Err: errors.New("must specify name")}
	}

	var count int64
	row := s.db.QueryRow(ctx, isEmptyNamespaceQuery, name)
	if err := row.Scan(&count); err != nil {
		if err == pgx.ErrNoRows {
			return false, &store.ErrNotFound{Key: namespaceStoreKey(name)}
		}
		return false, &store.ErrInternal{Message: err.Error()}
	}
	return count == 0, nil
}

func namespaceStoreKey(name string) string {
	return fmt.Sprintf("%s(name)=(%s)", entityConfigStoreName, name)
}
