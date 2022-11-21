package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/poll"
	"github.com/sensu/sensu-go/backend/selector"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
)

type StoreConfig struct {
	DB             DBI
	WatchInterval  time.Duration
	WatchTxnWindow time.Duration
}

func NewStore(cfg StoreConfig) *Store {
	return &Store{
		db:             cfg.DB,
		watchInterval:  cfg.WatchInterval,
		watchTxnWindow: cfg.WatchTxnWindow,
	}
}

type Store struct {
	db             DBI
	watchInterval  time.Duration
	watchTxnWindow time.Duration
}

func (s *Store) GetConfigStore() storev2.ConfigStore {
	return &ConfigStore{
		db:             s.db,
		watchTxnWindow: s.watchTxnWindow,
		watchInterval:  s.watchInterval,
	}
}

func (s *Store) GetEntityConfigStore() storev2.EntityConfigStore {
	return &EntityConfigStore{
		db: s.db,
	}
}

func (s *Store) GetEntityStateStore() storev2.EntityStateStore {
	return &EntityStateStore{
		db: s.db,
	}
}

func (s *Store) GetNamespaceStore() storev2.NamespaceStore {
	return &NamespaceStore{
		db: s.db,
	}
}

// legacy
func (s *Store) GetEventStore() store.EventStore {
	return &EventStore{
		db: s.db,
	}
}

// legacy
func (s *Store) GetEntityStore() store.EntityStore {
	return &EntityStore{
		db: s.db,
	}
}

const pgUniqueViolationCode = "23505"

type DBI interface {
	Begin(context.Context) (pgx.Tx, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

type ConfigStore struct {
	db             DBI
	watchInterval  time.Duration
	watchTxnWindow time.Duration
}

type configRecord struct {
	id          int64
	apiVersion  string
	apiType     string
	namespace   string
	name        string
	labels      string
	annotations []byte
	resource    string
	createdAt   sql.NullTime
	updatedAt   sql.NullTime
	deletedAt   sql.NullTime
}

type listTemplateValues struct {
	Limit       int64
	Offset      int64
	SelectorSQL string
}

func NewConfigStore(db DBI) *ConfigStore {
	return &ConfigStore{
		db: db,
	}
}

func (c *ConfigStore) GetEntityConfigStore() storev2.EntityConfigStore {
	return NewEntityConfigStore(c.db)
}

func (c *ConfigStore) GetEntityStateStore() storev2.EntityStateStore {
	return NewEntityStateStore(c.db)
}

func (c *ConfigStore) GetNamespaceStore() storev2.NamespaceStore {
	return NewNamespaceStore(c.db)
}

func (s *ConfigStore) Initialize(ctx context.Context, fn storev2.InitializeFunc) (err error) {
	tx, cerr := s.db.Begin(ctx)
	if cerr != nil {
		return cerr
	}
	defer tx.Commit(ctx)
	return fn(ctx, &Store{db: tx})
}

func (s *ConfigStore) GetPoller(req storev2.ResourceRequest) (poll.Table, error) {
	return newConfigurationPoller(req, s.db)
}

func (s *ConfigStore) Watch(ctx context.Context, req storev2.ResourceRequest) <-chan []storev2.WatchEvent {
	if req.APIVersion == "" || req.Type == "" {
		return nil
	}
	return NewWatcher(s, s.watchInterval, s.watchTxnWindow).Watch(ctx, req)
}

func (s *ConfigStore) CreateOrUpdate(ctx context.Context, request storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	if err := request.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	typeMeta, meta, jsonLabels, bytesAnnotations, jsonResource, err := extractResourceData(wrapper)
	if err != nil {
		return err
	}

	args := []interface{}{typeMeta.APIVersion, typeMeta.Type, meta.Namespace, meta.Name, jsonLabels, bytesAnnotations, jsonResource}

	tags, err := s.db.Exec(ctx, CreateOrUpdateConfigQuery, args...)
	if err != nil {
		return err
	}
	if tags.RowsAffected() == 0 {
		return errors.New("no rows inserted")
	}

	return nil
}

func (s *ConfigStore) UpdateIfExists(ctx context.Context, request storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	if err := request.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	typeMeta, meta, jsonLabels, bytesAnnotations, jsonResource, err := extractResourceData(wrapper)
	if err != nil {
		return err
	}

	args := []interface{}{jsonLabels, bytesAnnotations, jsonResource, typeMeta.APIVersion, typeMeta.Type, meta.Namespace, meta.Name}

	tags, err := s.db.Exec(ctx, UpdateIfExistsConfigQuery, args...)
	if err != nil {
		return err
	}
	if tags.RowsAffected() == 0 {
		return &store.ErrNotFound{Key: fmt.Sprintf("%s.%s/%s/%s", typeMeta.APIVersion, typeMeta.Type, request.Namespace, request.Name)}
	}

	return nil
}

func (s *ConfigStore) CreateIfNotExists(ctx context.Context, request storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	if err := request.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	typeMeta, meta, jsonLabels, bytesAnnotations, jsonResource, err := extractResourceData(wrapper)
	if err != nil {
		return err
	}

	args := []interface{}{typeMeta.APIVersion, typeMeta.Type, meta.Namespace, meta.Name, jsonLabels, bytesAnnotations, jsonResource}

	tags, err := s.db.Exec(ctx, CreateConfigIfNotExistsQuery, args...)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgUniqueViolationCode {
			return &store.ErrAlreadyExists{Key: fmt.Sprintf("%s.%s/%s/%s", typeMeta.APIVersion, typeMeta.Type, request.Namespace, request.Name)}
		}
		return err
	}
	if tags.RowsAffected() == 0 {
		return &store.ErrAlreadyExists{Key: fmt.Sprintf("%s.%s/%s/%s", typeMeta.APIVersion, typeMeta.Type, request.Namespace, request.Name)}
	}

	return nil
}

func (s *ConfigStore) Get(ctx context.Context, request storev2.ResourceRequest) (storev2.Wrapper, error) {
	if err := request.Validate(); err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}

	args := []interface{}{request.APIVersion, request.Type, request.Namespace, request.Name}

	row := s.db.QueryRow(ctx, GetConfigQuery, args...)
	var result configRecord

	if err := row.Scan(&result.id, &result.labels, &result.annotations, &result.resource, &result.createdAt, &result.updatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, &store.ErrNotFound{Key: fmt.Sprintf("%s.%s/%s/%s", request.APIVersion, request.Type, request.Namespace, request.Name)}
		}
		return nil, err
	}

	return &wrap.Wrapper{
		TypeMeta:    &corev2.TypeMeta{APIVersion: request.APIVersion, Type: request.Type},
		Encoding:    wrap.Encoding_json,
		Compression: 0,
		Value:       []byte(result.resource),
	}, nil
}

func (s *ConfigStore) Delete(ctx context.Context, request storev2.ResourceRequest) error {
	if err := request.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	args := []interface{}{request.APIVersion, request.Type, request.Namespace, request.Name}

	cmdTag, err := s.db.Exec(ctx, DeleteConfigQuery, args...)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return &store.ErrNotFound{Key: fmt.Sprintf("%s.%s/%s/%s", request.APIVersion, request.Type, request.Namespace, request.Name)}
	}

	return nil
}

func (s *ConfigStore) List(ctx context.Context, request storev2.ResourceRequest, predicate *store.SelectionPredicate) (storev2.WrapList, error) {
	if err := request.Validate(); err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}

	selectorSQL, selectorArgs, err := getSelectorSQL(ctx)
	if err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}

	tmpl, err := template.New("listResourceQuery").Parse(ListConfigQueryTmpl)
	if err != nil {
		return nil, err
	}
	var queryBuilder strings.Builder
	if predicate == nil {
		predicate = &store.SelectionPredicate{}
	}
	templValues := listTemplateValues{
		Limit:       predicate.Limit,
		Offset:      predicate.Offset,
		SelectorSQL: strings.TrimSpace(selectorSQL),
	}
	if err := tmpl.Execute(&queryBuilder, templValues); err != nil {
		return nil, err
	}

	args := []interface{}{request.APIVersion, request.Type, request.Namespace}
	args = append(args, selectorArgs...)

	query := queryBuilder.String()
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wrapList := make(wrap.List, 0)
	for rows.Next() {
		var dbResource configRecord
		err := rows.Scan(&dbResource.id, &dbResource.labels, &dbResource.annotations, &dbResource.resource, &dbResource.createdAt, &dbResource.updatedAt)
		if err != nil {
			return nil, err
		}

		wrapped := wrap.Wrapper{
			TypeMeta:    &corev2.TypeMeta{APIVersion: request.APIVersion, Type: request.Type},
			Encoding:    wrap.Encoding_json,
			Compression: wrap.Compression_none,
			Value:       []byte(dbResource.resource),
		}
		wrapList = append(wrapList, &wrapped)
	}

	return wrapList, nil
}

func (s *ConfigStore) Exists(ctx context.Context, request storev2.ResourceRequest) (bool, error) {
	if err := request.Validate(); err != nil {
		return false, &store.ErrNotValid{Err: err}
	}

	args := []interface{}{request.APIVersion, request.Type, request.Namespace, request.Name}

	row := s.db.QueryRow(ctx, ExistsConfigQuery, args...)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *ConfigStore) Patch(ctx context.Context, request storev2.ResourceRequest, wrapper storev2.Wrapper, patcher patch.Patcher, condition *store.ETagCondition) error {
	if err := request.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	key := storeKey(request)

	w, ok := wrapper.(*wrap.Wrapper)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("etcdstore only works with wrap.Wrapper, not %T", wrapper)}
	}

	resource, err := w.Unwrap()
	if err != nil {
		return err
	}

	patchedResource, err := patcher.Patch(w.Value)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(patchedResource, resource); err != nil {
		return err
	}

	if err := resource.Validate(); err != nil {
		return err
	}

	// Special case for entities; we need to make sure we keep the per-entity
	// subscription
	if e, ok := resource.(*corev3.EntityConfig); ok {
		e.Subscriptions = corev2.AddEntitySubscription(e.Metadata.Name, e.Subscriptions)
	}

	// Re-wrap the resource
	wrappedPatch, err := wrap.Resource(resource)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}
	*w = *wrappedPatch

	return s.UpdateIfExists(ctx, request, w)
}

func (s *ConfigStore) Count(ctx context.Context, request storev2.ResourceRequest) (int, error) {
	if err := request.Validate(); err != nil {
		return 0, &store.ErrNotValid{Err: err}
	}

	selectorSQL, selectorArgs, err := getSelectorSQL(ctx)
	if err != nil {
		return 0, &store.ErrNotValid{Err: err}
	}

	tmpl, err := template.New("countResourceQuery").Parse(CountConfigQueryTmpl)
	if err != nil {
		return 0, err
	}
	var queryBuilder strings.Builder
	templValues := listTemplateValues{
		SelectorSQL: strings.TrimSpace(selectorSQL),
	}
	if err := tmpl.Execute(&queryBuilder, templValues); err != nil {
		return 0, err
	}

	args := []interface{}{request.APIVersion, request.Type, request.Namespace}
	args = append(args, selectorArgs...)

	query := queryBuilder.String()
	row := s.db.QueryRow(ctx, query, args...)
	var ct int
	if err := row.Scan(&ct); err != nil {
		return 0, err
	}
	return ct, nil
}

func (s *ConfigStore) NamespaceStore() storev2.NamespaceStore {
	return NewNamespaceStore(s.db)
}

func (s *ConfigStore) EntityConfigStore() storev2.EntityConfigStore {
	return NewEntityConfigStore(s.db)
}

func (s *ConfigStore) EntityStateStore() storev2.EntityStateStore {
	return NewEntityStateStore(s.db)
}

func labelsToJSON(labels map[string]string) (string, error) {
	if len(labels) == 0 {
		return "{}", nil
	}

	bytes, err := json.Marshal(labels)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func annotationsToBytes(annotations map[string]string) ([]byte, error) {
	if len(annotations) == 0 {
		return []byte("{}"), nil
	}

	return json.Marshal(annotations)
}

func extractResourceData(wrapper storev2.Wrapper) (typeMeta corev2.TypeMeta, meta *corev2.ObjectMeta, jsonLabels string,
	bytesAnnotations, jsonResource []byte, err error) {
	res, err := wrapper.Unwrap()
	if err != nil {
		return
	}

	meta = res.GetMetadata()

	jsonLabels, err = labelsToJSON(meta.Labels)
	if err != nil {
		return
	}
	bytesAnnotations, err = annotationsToBytes(meta.Annotations)
	if err != nil {
		return
	}
	jsonResource, err = json.Marshal(res)
	if err != nil {
		return
	}

	resProxy := corev3.V2ResourceProxy{Resource: res}
	typeMeta = resProxy.GetTypeMeta()

	return
}

// StoreKey converts a ResourceRequest into a key that uniquely identifies a
// singular resource, or collection of resources, in a namespace.
func storeKey(req storev2.ResourceRequest) string {
	return store.NewKeyBuilder(req.StoreName).WithNamespace(req.Namespace).Build(req.Name)
}

func getSelectorSQL(ctx context.Context) (string, []interface{}, error) {
	ctxSelector := selector.SelectorFromContext(ctx)

	if ctxSelector != nil && len(ctxSelector.Operations) > 0 {
		argCounter := argCounter{value: 3}
		builder := NewConfigSelectorSQLBuilder(ctxSelector)
		return builder.GetSelectorCond(&argCounter)
	}

	return "", nil, nil
}

type configurationPoller struct {
	db  DBI
	req storev2.ResourceRequest
}

func (e *configurationPoller) Now(ctx context.Context) (time.Time, error) {
	var now time.Time
	row := e.db.QueryRow(ctx, "SELECT NOW();")
	if err := row.Scan(&now); err != nil {
		return now, &store.ErrInternal{Message: err.Error()}
	}
	return now, nil
}

func (e *configurationPoller) Since(ctx context.Context, updatedSince time.Time) ([]poll.Row, error) {
	queryParams := []interface{}{&e.req.APIVersion, &e.req.Type, &updatedSince}
	rows, err := e.db.Query(ctx, ConfigNotificationQuery, queryParams...)
	if err != nil {
		logger.Errorf("entity config since query failed with error %v", err)
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	defer rows.Close()
	var since []poll.Row
	for rows.Next() {
		var record configRecord
		if err := rows.Scan(&record.id, &record.apiVersion, &record.apiType, &record.namespace, &record.name,
			&record.resource, &record.createdAt, &record.updatedAt, &record.deletedAt); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		if err := rows.Err(); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		id := fmt.Sprintf("%s/%s", record.namespace, record.name)
		pollResult := poll.Row{
			Id: id,
			Resource: &wrap.Wrapper{
				TypeMeta:    &corev2.TypeMeta{APIVersion: e.req.APIVersion, Type: e.req.Type},
				Encoding:    wrap.Encoding_json,
				Compression: wrap.Compression_none,
				Value:       []byte(record.resource),
			},
			CreatedAt: record.createdAt.Time,
			UpdatedAt: record.updatedAt.Time,
		}
		if record.deletedAt.Valid {
			pollResult.DeletedAt = &record.deletedAt.Time
		}
		since = append(since, pollResult)
	}
	return since, nil
}

func newConfigurationPoller(req storev2.ResourceRequest, db DBI) (poll.Table, error) {
	return &configurationPoller{db: db, req: req}, nil
}
