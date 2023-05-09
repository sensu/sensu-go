package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/poll"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/memory"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"golang.org/x/time/rate"
)

type StoreConfig struct {
	DB                DBI
	MaxTPS            int
	WatchInterval     time.Duration
	WatchTxnWindow    time.Duration
	Bus               messaging.MessageBus
	DisableEventCache bool
}

func NewStore(cfg StoreConfig) *Store {
	return &Store{
		db:                cfg.DB,
		watchInterval:     cfg.WatchInterval,
		watchTxnWindow:    cfg.WatchTxnWindow,
		maxTPS:            cfg.MaxTPS,
		bus:               cfg.Bus,
		disableEventCache: cfg.DisableEventCache,
	}
}

type Store struct {
	db                DBI
	watchInterval     time.Duration
	watchTxnWindow    time.Duration
	eventStore        store.EventStore
	maxTPS            int
	once              sync.Once
	bus               messaging.MessageBus
	disableEventCache bool
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
	s.once.Do(func() {
		sstore := s.GetSilencesStore()
		eventStore, _ := NewEventStore(s.db, sstore, Config{})
		if s.disableEventCache {
			s.eventStore = eventStore
			return
		}
		cfg := memory.EventStoreConfig{
			BackingStore:    eventStore,
			FlushInterval:   time.Second,
			EventWriteLimit: rate.Limit(s.maxTPS),
			SilenceStore:    sstore,
			Bus:             s.bus,
		}
		memstore := memory.NewEventStore(cfg)
		memstore.Start(context.Background())
		s.eventStore = memstore
	})
	return s.eventStore
}

// legacy
func (s *Store) GetEntityStore() store.EntityStore {
	return &EntityStore{
		db: s.db,
	}
}

func (s *Store) GetSilencesStore() storev2.SilencesStore {
	return &SilenceStore{db: s.db}
}

const pgUniqueViolationCode = "23505"

type DBI interface {
	Begin(context.Context) (pgx.Tx, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	SendBatch(ctx context.Context, b *pgx.Batch) (br pgx.BatchResults)
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
	etag        storev2.ETag
}

type listTemplateValues struct {
	Limit       int64
	Offset      int64
	Namespaced  bool
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

	data, err := extractResourceData(wrapper)
	if err != nil {
		return err
	}

	meta, typeMeta := data.Metadata, data.TypeMeta
	args := []interface{}{typeMeta.APIVersion, typeMeta.Type, meta.Namespace, meta.Name, data.Labels, data.Annotations, data.Fields, data.Resource}
	query := CreateOrUpdateConfigQuery

	ifNoneMatch := storev2.IfNoneMatchFromContext(ctx)
	ifMatch := storev2.IfMatchFromContext(ctx)

	if ifMatch != nil {
		return &store.ErrNotValid{Err: errors.New("can't use IfMatch with this method")}
	}
	if ifNoneMatch != nil {
		return s.createOrUpdateIfNoneMatch(ctx, args, ifNoneMatch)
	}

	var (
		inserted       bool
		prevEtag, etag storev2.ETag
	)
	row := s.db.QueryRow(ctx, query, args...)
	if err := row.Scan(&inserted, &prevEtag, &etag); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	txInfo := storev2.TxInfoFromContext(ctx)
	if txInfo != nil {
		var record storev2.TxRecordInfo
		if inserted {
			record.Created = true
		} else {
			// it's only updated if the etag didn't change
			if !etag.Equals(prevEtag) {
				record.Updated = true
			}
			record.PrevETag = storev2.ETag(prevEtag)
		}
		record.ETag = storev2.ETag(etag)
		txInfo.Records = append(txInfo.Records, record)
	}

	return nil
}

func (s *ConfigStore) createOrUpdateIfNoneMatch(ctx context.Context, args []interface{}, ifNoneMatch storev2.IfNoneMatch) error {
	var (
		inserted       bool
		prevEtag, etag storev2.ETag
	)
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	defer tx.Commit(ctx)
	row := tx.QueryRow(ctx, GetETagOnlyQuery, args[:4]...)
	if err := row.Scan(&etag); err != nil {
		if err == pgx.ErrNoRows {
			goto OK
		}
		return &store.ErrInternal{Message: err.Error()}
	}
	for _, tag := range ifNoneMatch {
		if etag.Equals(storev2.ETag(tag)) {
			return storev2.ErrPreconditionFailed
		}
	}
OK:
	row = tx.QueryRow(ctx, CreateOrUpdateConfigQuery, args...)
	if err := row.Scan(&inserted, &prevEtag, &etag); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	txInfo := storev2.TxInfoFromContext(ctx)
	if txInfo != nil {
		var record storev2.TxRecordInfo
		if inserted {
			record.Created = true
		} else {
			if !etag.Equals(prevEtag) {
				// it's only updated if the etag didn't change
				record.Updated = true
			}
			record.PrevETag = storev2.ETag(prevEtag)
		}
		record.ETag = storev2.ETag(etag)
		txInfo.Records = append(txInfo.Records, record)
	}
	return nil
}

func (s *ConfigStore) UpdateIfExists(ctx context.Context, request storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	if err := request.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	data, err := extractResourceData(wrapper)
	if err != nil {
		return err
	}

	meta, typeMeta := data.Metadata, data.TypeMeta
	args := []interface{}{data.Labels, data.Annotations, data.Fields, data.Resource, typeMeta.APIVersion, typeMeta.Type, meta.Namespace, meta.Name}
	query := UpdateIfExistsConfigQuery

	ifMatch := storev2.IfMatchFromContext(ctx)
	ifNoneMatch := storev2.IfNoneMatchFromContext(ctx)

	var etag, prevEtag storev2.ETag

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	defer tx.Commit(ctx)

	getArgs := []interface{}{typeMeta.APIVersion, typeMeta.Type, meta.Namespace, meta.Name}

	row := tx.QueryRow(ctx, GetETagOnlyQuery, getArgs...)
	if err := row.Scan(&prevEtag); err != nil {
		if err == pgx.ErrNoRows {
			return &store.ErrNotFound{Key: fmt.Sprintf("%s.%s/%s/%s", request.APIVersion, request.Type, request.Namespace, request.Name)}
		}
		return &store.ErrInternal{Message: err.Error()}
	}

	if ifMatch != nil {
		for _, tag := range ifMatch {
			if prevEtag.Equals(tag) {
				goto MATCHED
			}
		}
		return storev2.ErrPreconditionFailed
	}

	for _, tag := range ifNoneMatch {
		if prevEtag.Equals(tag) {
			return storev2.ErrPreconditionFailed
		}
	}

MATCHED:

	row = tx.QueryRow(ctx, query, args...)
	if err := row.Scan(&etag); err != nil {
		// impossible because the previous part of the tx succeeded, but
		// handle it anyways just in case something changes down the road.
		if err == pgx.ErrNoRows {
			if ifNoneMatch != nil || ifMatch != nil {
				return storev2.ErrPreconditionFailed
			}
			return &store.ErrNotFound{Key: fmt.Sprintf("%s.%s/%s/%s", request.APIVersion, request.Type, request.Namespace, request.Name)}
		}
		return err
	}
	txInfo := storev2.TxInfoFromContext(ctx)
	if txInfo != nil {
		var record storev2.TxRecordInfo
		record.Updated = true
		record.PrevETag = storev2.ETag(prevEtag)
		record.ETag = storev2.ETag(etag)
		txInfo.Records = append(txInfo.Records, record)
	}

	return nil
}

func (s *ConfigStore) CreateIfNotExists(ctx context.Context, request storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	if err := request.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	data, err := extractResourceData(wrapper)
	if err != nil {
		return err
	}

	meta, typeMeta := data.Metadata, data.TypeMeta
	args := []interface{}{typeMeta.APIVersion, typeMeta.Type, meta.Namespace, meta.Name, data.Labels, data.Annotations, data.Fields, data.Resource}

	row := s.db.QueryRow(ctx, CreateConfigIfNotExistsQuery, args...)
	var etag []byte

	if err := row.Scan(&etag); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == pgUniqueViolationCode {
			return &store.ErrAlreadyExists{Key: fmt.Sprintf("%s.%s/%s/%s", typeMeta.APIVersion, typeMeta.Type, request.Namespace, request.Name)}
		}
		return err
	}
	txInfo := storev2.TxInfoFromContext(ctx)
	if txInfo != nil {
		var record storev2.TxRecordInfo
		record.Created = true
		record.ETag = storev2.ETag(etag)
		txInfo.Records = append(txInfo.Records, record)
	}

	return nil
}

func (s *ConfigStore) Get(ctx context.Context, request storev2.ResourceRequest) (storev2.Wrapper, error) {
	if err := request.Validate(); err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}

	query := GetConfigQuery
	args := []interface{}{request.APIVersion, request.Type, request.Namespace, request.Name}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	defer tx.Commit(ctx)

	if err := s.checkGetPreconditions(ctx, tx, args); err != nil {
		return nil, err
	}

	row := tx.QueryRow(ctx, query, args...)
	var rec configRecord

	if err := row.Scan(&rec.id, &rec.labels, &rec.annotations, &rec.resource, &rec.createdAt, &rec.updatedAt, &rec.etag); err != nil {
		if err == pgx.ErrNoRows {
			return nil, &store.ErrNotFound{Key: fmt.Sprintf("%s.%s/%s/%s", request.APIVersion, request.Type, request.Namespace, request.Name)}
		}
		return nil, &store.ErrInternal{Message: err.Error()}
	}

	return &wrap.Wrapper{
		TypeMeta:    &corev2.TypeMeta{APIVersion: request.APIVersion, Type: request.Type},
		Encoding:    wrap.Encoding_json,
		Compression: 0,
		Value:       []byte(rec.resource),
		CreatedAt:   rec.createdAt.Time,
		UpdatedAt:   rec.updatedAt.Time,
		ETag:        rec.etag.String(),
	}, nil
}

func (s *ConfigStore) checkGetPreconditions(ctx context.Context, tx pgx.Tx, args []interface{}) error {
	ifMatch := storev2.IfMatchFromContext(ctx)
	ifNoneMatch := storev2.IfNoneMatchFromContext(ctx)

	if ifMatch == nil && ifNoneMatch == nil {
		return nil
	}

	var etag storev2.ETag

	row := tx.QueryRow(ctx, GetETagOnlyQuery, args...)
	if err := row.Scan(&etag); err != nil {
		if err == pgx.ErrNoRows {
			return &store.ErrNotFound{Key: fmt.Sprintf("%s.%s/%s/%s", args...)}
		}
	}
	if ifMatch != nil {
		for _, tag := range ifMatch {
			if etag.Equals(tag) {
				return nil
			}
		}
		return storev2.ErrPreconditionFailed
	}

	for _, tag := range ifNoneMatch {
		if etag.Equals(tag) {
			return storev2.ErrPreconditionFailed
		}
	}

	return nil
}

func (s *ConfigStore) Delete(ctx context.Context, request storev2.ResourceRequest) error {
	if err := request.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	query := DeleteConfigQuery
	args := []interface{}{request.APIVersion, request.Type, request.Namespace, request.Name}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	defer tx.Commit(ctx)

	if err := s.checkGetPreconditions(ctx, tx, args); err != nil {
		return err
	}

	cmdTag, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return &store.ErrNotFound{Key: fmt.Sprintf("%s.%s/%s/%s", request.APIVersion, request.Type, request.Namespace, request.Name)}
	}

	txInfo := storev2.TxInfoFromContext(ctx)
	if txInfo != nil {
		var record storev2.TxRecordInfo
		record.Deleted = true
		txInfo.Records = append(txInfo.Records, record)
	}

	return nil
}

func (s *ConfigStore) List(ctx context.Context, request storev2.ResourceRequest, pred *store.SelectionPredicate) (storev2.WrapList, error) {
	if err := request.Validate(); err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}

	selectorSQL, selectorArgs, err := getSelectorSQL(ctx, request.APIVersion, request.Type, 5)
	if err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}

	limit, offset, err := getLimitAndOffset(pred)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("listResourceQuery").Parse(ListConfigQueryTmpl)
	if err != nil {
		return nil, err
	}
	var queryBuilder strings.Builder
	if pred == nil {
		pred = &store.SelectionPredicate{}
	}

	templValues := listTemplateValues{
		Offset:      offset,
		SelectorSQL: strings.TrimSpace(selectorSQL),
		Namespaced:  request.Namespace != "",
	}
	if limit.Valid {
		templValues.Limit = limit.Int64
	}
	if err := tmpl.Execute(&queryBuilder, templValues); err != nil {
		return nil, err
	}

	if pred.UpdatedSince == "" {
		pred.UpdatedSince = beginningOfTime
	}

	var updatedSince time.Time
	if err := updatedSince.UnmarshalText([]byte(pred.UpdatedSince)); err != nil {
		return nil, &store.ErrNotValid{Err: fmt.Errorf("bad UpdatedSince time: %s", err)}
	}

	args := []interface{}{request.APIVersion, request.Type, request.Namespace, updatedSince, pred.IncludeDeletes}
	args = append(args, selectorArgs...)

	query := queryBuilder.String()
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wrapList := make(wrap.List, 0)
	for rows.Next() {
		var rec configRecord
		err := rows.Scan(&rec.id, &rec.labels, &rec.annotations, &rec.resource, &rec.createdAt, &rec.updatedAt, &rec.deletedAt, &rec.etag)
		if err != nil {
			return nil, err
		}

		wrapped := wrap.Wrapper{
			TypeMeta:    &corev2.TypeMeta{APIVersion: request.APIVersion, Type: request.Type},
			Encoding:    wrap.Encoding_json,
			Compression: wrap.Compression_none,
			Value:       []byte(rec.resource),
			CreatedAt:   rec.createdAt.Time,
			UpdatedAt:   rec.updatedAt.Time,
			DeletedAt:   rec.deletedAt.Time,
			ETag:        rec.etag.String(),
		}
		wrapList = append(wrapList, &wrapped)
	}
	if pred != nil {
		if int64(len(wrapList)) < pred.Limit {
			pred.Continue = ""
		}
	}

	return wrapList, nil
}

func (s *ConfigStore) Exists(ctx context.Context, request storev2.ResourceRequest) (bool, error) {
	if err := request.Validate(); err != nil {
		return false, &store.ErrNotValid{Err: err}
	}

	query := ExistsConfigQuery
	args := []interface{}{request.APIVersion, request.Type, request.Namespace, request.Name}

	ifMatch := storev2.IfMatchFromContext(ctx)
	ifNoneMatch := storev2.IfNoneMatchFromContext(ctx)

	if ifMatch != nil {
		query = ExistsConfigIfMatchQuery
		args = append(args, ifMatch)
	} else if ifNoneMatch != nil {
		query = ExistsConfigIfNoneMatchQuery
		args = append(args, ifNoneMatch)
	}

	row := s.db.QueryRow(ctx, query, args...)

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
		return &store.ErrNotValid{Err: fmt.Errorf("store only works with wrap.Wrapper, not %T", wrapper)}
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

	selectorSQL, selectorArgs, err := getSelectorSQL(ctx, request.APIVersion, request.Type, 3)
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
		Namespaced:  request.Namespace != "",
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

func extractResourceData(wrapper storev2.Wrapper) (data resourceData, err error) {
	res, err := wrapper.Unwrap()
	if err != nil {
		return
	}

	data.Metadata = res.GetMetadata()

	data.Labels, err = labelsToJSON(data.Metadata.Labels)
	if err != nil {
		return
	}

	data.Annotations, err = annotationsToBytes(data.Metadata.Annotations)
	if err != nil {
		return
	}

	data.Fields = []byte("{}")
	if fielder, ok := res.(corev3.Fielder); ok {
		data.Fields, _ = json.Marshal(fielder.Fields())
	}

	data.Resource, err = json.Marshal(res)
	if err != nil {
		return
	}

	resProxy := corev3.V2ResourceProxy{Resource: res}
	data.TypeMeta = resProxy.GetTypeMeta()

	return
}

// StoreKey converts a ResourceRequest into a key that uniquely identifies a
// singular resource, or collection of resources, in a namespace.
func storeKey(req storev2.ResourceRequest) string {
	return store.NewKeyBuilder(req.StoreName).WithNamespace(req.Namespace).Build(req.Name)
}

func getSelectorSQL(ctx context.Context, apiVersion, typeName string, nargs int) (string, []interface{}, error) {
	ctxSelector := storev2.SelectorFromContext(ctx, corev2.TypeMeta{APIVersion: apiVersion, Type: typeName})

	if ctxSelector != nil && len(ctxSelector.Operations) > 0 {
		argCounter := argCounter{value: nargs}
		builder := NewConfigSelectorSQLBuilder(ctxSelector)
		return builder.GetSelectorCond(&argCounter)
	}

	return "", nil, nil
}

type resourceData struct {
	TypeMeta    corev2.TypeMeta
	Metadata    *corev2.ObjectMeta
	Labels      string
	Annotations []byte
	Fields      []byte
	Resource    []byte
}
