package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	json "github.com/json-iterator/go"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/poll"
	"github.com/sensu/sensu-go/backend/seeds"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/util/retry"
)

type StoreV2 struct {
	db *pgxpool.Pool

	watchInterval  time.Duration
	watchTxnWindow time.Duration
}

func NewStoreV2(pg *pgxpool.Pool) *StoreV2 {
	return &StoreV2{
		db: pg,
	}
}

type crudOp int

const (
	createOrUpdateQ    crudOp = 0
	updateIfExistsQ    crudOp = 1
	createIfNotExistsQ crudOp = 2
	getQ               crudOp = 3
	deleteQ            crudOp = 4
	listQ              crudOp = 5
	listAllQ           crudOp = 6
	existsQ            crudOp = 7
	patchQ             crudOp = 8
	getMultipleQ       crudOp = 9
	hardDeleteQ        crudOp = 10
	hardDeletedQ       crudOp = 11
)

type wrapperWithParams interface {
	storev2.Wrapper
	SQLParams() []any
}

type namedWrapperWithParams interface {
	wrapperWithParams
	GetName() string
}

type wrapperWithStatus interface {
	storev2.Wrapper
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetDeletedAt() sql.NullTime
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
	SetDeletedAt(sql.NullTime)
}

type namespacedResourceNames map[string][]string

type resourceRequestWrapperMap map[storev2.ResourceRequest]storev2.Wrapper

type ErrStoreV2 struct {
	Method    string
	StoreName string
	Err       error
}

func (e ErrStoreV2) Error() string {
	return fmt.Sprintf("%s %s: %s", e.Method, e.StoreName, e.Err)
}

func (e ErrStoreV2) Unwrap() error {
	return e.Err
}

func newStoreV2Error(method, storeName string, err error) error {
	if err != nil {
		return ErrStoreV2{
			Method:    method,
			StoreName: storeName,
			Err:       err,
		}
	} else {
		return nil
	}
}

func (s *StoreV2) lookupWrapper(req storev2.ResourceRequest, op crudOp) namedWrapperWithParams {
	storeName := req.StoreName
	switch storeName {
	default:
		msg := fmt.Sprintf("no wrapper type defined for store: %s", storeName)
		panic(msg)
	}
}

func (s *StoreV2) lookupQuery(req storev2.ResourceRequest, op crudOp) (string, bool) {
	storeName := req.StoreName
	//ordering := req.SortOrder
	switch storeName {
	default:
		return "", false
	}
}

func errNoQuery(storeName string) error {
	return &store.ErrNotValid{
		Err: fmt.Errorf("postgres storage requested for %s, but not supported", storeName),
	}
}

func (s *StoreV2) CreateOrUpdate(ctx context.Context, req storev2.ResourceRequest, wrapper storev2.Wrapper) (fErr error) {
	defer func() {
		fErr = newStoreV2Error("CreateOrUpdate", req.StoreName, fErr)
	}()

	pwrap, ok := wrapper.(wrapperWithParams)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("bad wrapper type for postgres: %T", wrapper)}
	}

	query, ok := s.lookupQuery(req, createOrUpdateQ)
	if !ok {
		return errNoQuery(req.StoreName)
	}

	params := pwrap.SQLParams()

	if _, err := s.db.Exec(ctx, query, params...); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

func (s *StoreV2) UpdateIfExists(ctx context.Context, req storev2.ResourceRequest, wrapper storev2.Wrapper) (fErr error) {
	defer func() {
		fErr = newStoreV2Error("UpdateIfExists", req.StoreName, fErr)
	}()

	pwrap, ok := wrapper.(wrapperWithParams)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("bad wrapper type for postgres: %T", wrapper)}
	}

	query, ok := s.lookupQuery(req, updateIfExistsQ)
	if !ok {
		return errNoQuery(req.StoreName)
	}

	params := pwrap.SQLParams()

	rows, err := s.db.Query(ctx, query, params...)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	defer rows.Close()

	rowCount := 0
	for rows.Next() {
		rowCount++
	}
	if rowCount == 0 {
		return &store.ErrNotFound{Key: fmt.Sprintf("%s.%s", req.Namespace, req.Name)}
	}

	return nil
}

func (s *StoreV2) CreateIfNotExists(ctx context.Context, req storev2.ResourceRequest, wrapper storev2.Wrapper) (fErr error) {
	defer func() {
		fErr = newStoreV2Error("CreateIfNotExists", req.StoreName, fErr)
	}()

	pwrap, ok := wrapper.(wrapperWithParams)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("bad wrapper type for postgres: %T", wrapper)}
	}

	query, ok := s.lookupQuery(req, createIfNotExistsQ)
	if !ok {
		return errNoQuery(req.StoreName)
	}

	params := pwrap.SQLParams()

	if _, err := s.db.Exec(ctx, query, params...); err != nil {
		pgError, ok := err.(*pgconn.PgError)
		if ok {
			switch pgError.ConstraintName {
			case "entity_state_unique":
				return &store.ErrAlreadyExists{Key: fmt.Sprintf("%s/%s", req.Namespace, req.Name)}
			}
		}
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

func (s *StoreV2) Get(ctx context.Context, req storev2.ResourceRequest) (fWrapper storev2.Wrapper, fErr error) {
	defer func() {
		fErr = newStoreV2Error("Get", req.StoreName, fErr)
	}()

	op := getQ
	query, ok := s.lookupQuery(req, op)
	if !ok {
		return nil, errNoQuery(req.StoreName)
	}

	row := s.db.QueryRow(ctx, query, req.Namespace, req.Name)
	wrapper := s.lookupWrapper(req, op)
	if err := row.Scan(wrapper.SQLParams()...); err != nil {
		if err == pgx.ErrNoRows {
			return nil, &store.ErrNotFound{Key: fmt.Sprintf("%s.%s", req.Namespace, req.Name)}
		}
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	return wrapper, nil
}

func (s *StoreV2) GetMultiple(ctx context.Context, reqs []storev2.ResourceRequest) (wrapperMap resourceRequestWrapperMap, fErr error) {
	if len(reqs) == 0 {
		return resourceRequestWrapperMap{}, nil
	}

	storeName := reqs[0].StoreName
	defer func() {
		fErr = newStoreV2Error("GetMultiple", storeName, fErr)
	}()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Commit(ctx); err != nil && err != pgx.ErrTxClosed {
			logger.WithError(err).Error("error committing transaction for GetMultiple()")
		}
	}()

	type requestKey struct {
		Namespace string
		Name      string
	}

	keyedResourceRequests := map[requestKey]storev2.ResourceRequest{}
	namespacedResources := map[string][]string{}
	for _, req := range reqs {
		if err := req.Validate(); err != nil {
			return nil, &store.ErrNotValid{Err: err}
		}
		if storeName != req.StoreName {
			panic("GetMultiple() does not support requests with differing store names")
		}

		key := requestKey{
			Namespace: req.Namespace,
			Name:      req.Name,
		}
		keyedResourceRequests[key] = req
		namespacedResources[req.Namespace] = append(namespacedResources[req.Namespace], req.Name)
	}

	wrappers := resourceRequestWrapperMap{}
	op := getMultipleQ
	for namespace, resourceNames := range namespacedResources {
		query, ok := s.lookupQuery(reqs[0], op)
		if !ok {
			return nil, errNoQuery(storeName)
		}

		rows, err := tx.Query(ctx, query, namespace, resourceNames)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			wrapper := s.lookupWrapper(reqs[0], op)

			if err := rows.Scan(wrapper.SQLParams()...); err != nil {
				if err == pgx.ErrNoRows {
					continue
				}
				return nil, &store.ErrInternal{Message: err.Error()}
			}

			key := requestKey{
				Namespace: namespace,
				Name:      wrapper.GetName(),
			}
			req, ok := keyedResourceRequests[key]
			if !ok {
				panic("keyedResourceRequests key does not exist")
			}

			wrappers[req] = wrapper
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction for GetMultiple()")
	}

	return wrappers, nil
}

func (s *StoreV2) Delete(ctx context.Context, req storev2.ResourceRequest) (fErr error) {
	defer func() {
		fErr = newStoreV2Error("Delete", req.StoreName, fErr)
	}()

	query, ok := s.lookupQuery(req, deleteQ)
	if !ok {
		return errNoQuery(req.StoreName)
	}

	result, err := s.db.Exec(ctx, query, req.Namespace, req.Name)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	affected := result.RowsAffected()
	if affected < 1 {
		return &store.ErrNotFound{Key: fmt.Sprintf("%s.%s", req.Namespace, req.Name)}
	}
	return nil
}

func (s *StoreV2) HardDelete(ctx context.Context, req storev2.ResourceRequest) (fErr error) {
	defer func() {
		fErr = newStoreV2Error("HardDelete", req.StoreName, fErr)
	}()

	query, ok := s.lookupQuery(req, hardDeleteQ)
	if !ok {
		return errNoQuery(req.StoreName)
	}

	result, err := s.db.Exec(ctx, query, req.Namespace, req.Name)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	affected := result.RowsAffected()
	if affected < 1 {
		return &store.ErrNotFound{Key: fmt.Sprintf("%s.%s", req.Namespace, req.Name)}
	}
	return nil
}

func (s *StoreV2) List(ctx context.Context, req storev2.ResourceRequest, pred *store.SelectionPredicate) (list storev2.WrapList, fErr error) {
	defer func() {
		fErr = newStoreV2Error("List", req.StoreName, fErr)
	}()

	op := listQ
	query, ok := s.lookupQuery(req, op)
	if !ok {
		return nil, errNoQuery(req.StoreName)
	}
	limit, offset, err := getLimitAndOffset(pred)
	if err != nil {
		return nil, err
	}

	var namespace sql.NullString
	if req.Namespace != "" {
		namespace.String = req.Namespace
		namespace.Valid = true
	}

	rows, rerr := s.db.Query(ctx, query, namespace, limit, offset)
	if rerr != nil {
		return nil, &store.ErrInternal{Message: rerr.Error()}
	}
	defer rows.Close()
	wrapList := WrapList{}
	var rowCount int64
	for rows.Next() {
		rowCount++
		wrapper := s.lookupWrapper(req, op)
		if err := rows.Scan(wrapper.SQLParams()...); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		if err := rows.Err(); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		wrapList = append(wrapList, wrapper)
	}
	if pred != nil {
		if rowCount < pred.Limit {
			pred.Continue = ""
		}
	}
	return wrapList, nil
}

func (s *StoreV2) Exists(ctx context.Context, req storev2.ResourceRequest) (fExists bool, fErr error) {
	defer func() {
		fErr = newStoreV2Error("Exists", req.StoreName, fErr)
	}()

	query, ok := s.lookupQuery(req, existsQ)
	if !ok {
		return false, errNoQuery(req.StoreName)
	}

	row := s.db.QueryRow(ctx, query, req.Namespace, req.Name)
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

func (s *StoreV2) HardDeleted(ctx context.Context, req storev2.ResourceRequest) (fDeleted bool, fErr error) {
	defer func() {
		fErr = newStoreV2Error("HardDeleted", req.StoreName, fErr)
	}()

	op := hardDeletedQ
	query, ok := s.lookupQuery(req, op)
	if !ok {
		return false, errNoQuery(req.StoreName)
	}

	row := s.db.QueryRow(ctx, query, req.Namespace, req.Name)
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

func (s *StoreV2) Patch(ctx context.Context, req storev2.ResourceRequest, w storev2.Wrapper, patcher patch.Patcher, conditions *store.ETagCondition) (fErr error) {
	defer func() {
		fErr = newStoreV2Error("Patch", req.StoreName, fErr)
	}()

	if err := req.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	op := getQ
	query, ok := s.lookupQuery(req, op)
	if !ok {
		return errNoQuery(req.StoreName)
	}

	originalWrapper := s.lookupWrapper(req, op)

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

	row := tx.QueryRow(ctx, query, req.Namespace, req.Name)
	if err := row.Scan(originalWrapper.SQLParams()...); err != nil {
		if err == pgx.ErrNoRows {
			return &store.ErrNotFound{Key: fmt.Sprintf("%s.%s", req.Namespace, req.Name)}
		}
		return &store.ErrInternal{Message: err.Error()}
	}

	resource, err := originalWrapper.Unwrap()
	if err != nil {
		return &store.ErrDecode{Key: etcdstore.StoreKey(req), Err: err}
	}

	// Now determine the etag for the stored resource
	etag, err := store.ETag(resource)
	if err != nil {
		return err
	}

	if conditions != nil {
		if !store.CheckIfMatch(conditions.IfMatch, etag) {
			return &store.ErrPreconditionFailed{Key: etcdstore.StoreKey(req)}
		}
		if !store.CheckIfNoneMatch(conditions.IfNoneMatch, etag) {
			return &store.ErrPreconditionFailed{Key: etcdstore.StoreKey(req)}
		}
	}

	// Encode the stored resource to the JSON format
	original, err := json.Marshal(resource)
	if err != nil {
		return err
	}

	// Apply the patch to our original document (stored resource)
	patchedResource, err := patcher.Patch(original)
	if err != nil {
		return err
	}

	// Decode the resulting JSON document back into our resource
	if err := json.Unmarshal(patchedResource, &resource); err != nil {
		return err
	}

	// Validate the resource
	if err := resource.Validate(); err != nil {
		return err
	}

	wrapped, err := wrapWithPostgres(resource)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}

	pwrap, ok := wrapped.(wrapperWithParams)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("bad wrapper type for postgres: %T", wrapped)}
	}

	query, ok = s.lookupQuery(req, createOrUpdateQ)
	if !ok {
		return errNoQuery(req.StoreName)
	}

	if _, err := tx.Exec(ctx, query, pwrap.SQLParams()...); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

func (s *StoreV2) Watch(ctx context.Context, req storev2.ResourceRequest) <-chan []storev2.WatchEvent {
	eventChan := make(chan []storev2.WatchEvent, 32)

	var table poll.Table

	tableFactory, ok := getWatchStoreOverride(req.StoreName)
	if !ok {
		// assume configuration store?
		tableFactory = func(storev2.ResourceRequest, *pgxpool.Pool) (poll.Table, error) {
			return nil, fmt.Errorf("default watcher not yet implemented")
		}
	}
	table, err := tableFactory(req, s.db)
	if err != nil {
		panic(fmt.Errorf("could not create watcher for request %v: %v", req, err))
	}

	interval, txnWindow := s.watchInterval, s.watchTxnWindow
	if interval <= 0 {
		interval = time.Second
	}
	if txnWindow <= 0 {
		txnWindow = 5 * time.Second
	}
	poller := &poll.Poller{
		Interval:  interval,
		TxnWindow: txnWindow,
		Table:     table,
	}

	backoff := retry.ExponentialBackoff{
		Ctx: ctx,
	}
	err = backoff.Retry(func(retry int) (bool, error) {
		if err := poller.Initialize(ctx); err != nil {
			logger.Errorf("watcher initialize polling error on retry %d: %v", retry, err)
			return false, err
		}
		return true, nil
	})
	if err != nil {
		logger.Errorf("watcher failed to start: %v", err)
		close(eventChan)
		return eventChan
	}

	go s.watchLoop(ctx, req, poller, eventChan)
	return eventChan
}

func (s *StoreV2) watchLoop(ctx context.Context, req storev2.ResourceRequest, poller *poll.Poller, watchChan chan []storev2.WatchEvent) {
	defer close(watchChan)
	for {
		changes, err := poller.Next(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Error(err)
		}
		if len(changes) == 0 {
			continue
		}
		notifications := make([]storev2.WatchEvent, len(changes))
		for i, change := range changes {
			wrapper, ok := change.Resource.(storev2.Wrapper)
			if !ok {
				// Poller table must return Resource of type Wrapper.
				panic("postgres store watcher resource is not storev2.Wrapper")
			}
			notifications[i].Value = wrapper
			r, err := notifications[i].Value.Unwrap()
			if err != nil {
				notifications[i].Err = err
			}
			if r != nil {
				meta := r.GetMetadata()
				notifications[i].Key = storev2.ResourceRequest{
					Namespace: meta.Namespace,
					Name:      meta.Name,
					StoreName: r.StoreName(),
				}
			}
			switch change.Change {
			case poll.Create:
				notifications[i].Type = storev2.WatchCreate
			case poll.Update:
				notifications[i].Type = storev2.WatchUpdate
			case poll.Delete:
				notifications[i].Type = storev2.WatchDelete
			}
		}
		var status string
		select {
		case watchChan <- notifications:
			status = storev2.WatchEventsStatusHandled
		default:
			status = storev2.WatchEventsStatusDropped
		}
		storev2.WatchEventsProcessed.WithLabelValues(
			status,
			req.StoreName,
			req.Namespace,
			storev2.WatcherProviderPG,
		).Add(float64(len(notifications)))
	}
}

func (s *StoreV2) NamespaceStore() storev2.NamespaceStore {
	return NewNamespaceStore(s.db)
}

func (s *StoreV2) EntityConfigStore() storev2.EntityConfigStore {
	return NewEntityConfigStore(s.db)
}

func (s *StoreV2) EntityStateStore() storev2.EntityStateStore {
	return NewEntityStateStore(s.db)
}

func (s *StoreV2) Initialize(ctx context.Context, fn storev2.InitializeFunc) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Commit(ctx); err != nil && err != pgx.ErrTxClosed {
			logger.WithError(err).Error("error committing transaction for Initialize()")
		}
	}()

	initialized, err := s.isInitialized(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to check if cluster has been initialized: %w", err)
	}
	if initialized {
		logger.Info("store already initialized")
		return seeds.ErrAlreadyInitialized
	}

	if err := fn(ctx); err != nil {
		return err
	}
	return s.flagAsInitialized(ctx, tx)
}

func (s *StoreV2) isInitialized(ctx context.Context, tx pgx.Tx) (bool, error) {
	var found bool
	row := tx.QueryRow(ctx, isInitializedQuery)
	if err := row.Scan(&found); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, &store.ErrInternal{Message: err.Error()}
	}
	return found, nil
}

func (s *StoreV2) flagAsInitialized(ctx context.Context, tx pgx.Tx) error {
	if _, err := s.db.Exec(ctx, flagAsInitializedQuery); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	return nil
}

func wrapWithPostgres(resource corev3.Resource, opts ...wrap.Option) (storev2.Wrapper, error) {
	switch value := resource.(type) {
	case *corev3.EntityState:
		return WrapEntityState(value), nil
	default:
		return storev2.WrapResource(resource, opts...)
	}
}
