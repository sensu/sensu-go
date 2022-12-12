package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	json "github.com/json-iterator/go"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type StoreV2 struct {
	db          *pgxpool.Pool
	etcdStoreV2 storev2.Interface
}

func NewStoreV2(pg *pgxpool.Pool, client *clientv3.Client) *StoreV2 {
	return &StoreV2{
		db:          pg,
		etcdStoreV2: etcdstore.NewStore(client),
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
)

var (
	entityStateStoreName = new(corev3.EntityState).StoreName()
)

type wrapperParams interface {
	SQLParams() []interface{}
}

func (s *StoreV2) lookupQuery(req storev2.ResourceRequest, op crudOp) (string, bool) {
	storeName := req.StoreName
	ordering := req.SortOrder
	switch storeName {
	case entityStateStoreName:
		switch op {
		case createOrUpdateQ:
			return createOrUpdateEntityStateQuery, true
		case updateIfExistsQ:
			return updateIfExistsEntityStateQuery, true
		case createIfNotExistsQ:
			return createIfNotExistsEntityStateQuery, true
		case getQ:
			return getEntityStateQuery, true
		case deleteQ:
			return deleteEntityStateQuery, true
		case listQ, listAllQ:
			switch ordering {
			case storev2.SortDescend:
				return listEntityStateDescQuery, true
			default:
				return listEntityStateQuery, true
			}
		case existsQ:
			return existsEntityStateQuery, true
		case patchQ:
			return patchEntityStateQuery, true
		default:
			return "", false
		}
	default:
		return "", false
	}
}

func errNoQuery(storeName string) error {
	return &store.ErrNotValid{
		Err: fmt.Errorf("postgres storage requested for %s, but not supported", storeName),
	}
}

func (s *StoreV2) CreateOrUpdate(req storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	if !req.UsePostgres {
		return s.etcdStoreV2.CreateOrUpdate(req, wrapper)
	}
	pwrap, ok := wrapper.(wrapperParams)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("bad wrapper type for postgres: %T", wrapper)}
	}

	query, ok := s.lookupQuery(req, createOrUpdateQ)
	if !ok {
		return errNoQuery(req.StoreName)
	}

	params := pwrap.SQLParams()

	if _, err := s.db.Exec(req.Context, query, params...); err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

func (s *StoreV2) UpdateIfExists(req storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	if !req.UsePostgres {
		return s.etcdStoreV2.UpdateIfExists(req, wrapper)
	}
	pwrap, ok := wrapper.(wrapperParams)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("bad wrapper type for postgres: %T", wrapper)}
	}

	query, ok := s.lookupQuery(req, updateIfExistsQ)
	if !ok {
		return errNoQuery(req.StoreName)
	}

	params := pwrap.SQLParams()

	rows, err := s.db.Query(req.Context, query, params...)
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

func (s *StoreV2) CreateIfNotExists(req storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	if !req.UsePostgres {
		return s.etcdStoreV2.CreateIfNotExists(req, wrapper)
	}
	pwrap, ok := wrapper.(wrapperParams)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("bad wrapper type for postgres: %T", wrapper)}
	}

	query, ok := s.lookupQuery(req, createIfNotExistsQ)
	if !ok {
		return errNoQuery(req.StoreName)
	}

	params := pwrap.SQLParams()

	if _, err := s.db.Exec(req.Context, query, params...); err != nil {
		pgError, ok := err.(*pgconn.PgError)
		if ok {
			if pgError.ConstraintName == "entities_namespace_name_key" {
				return &store.ErrAlreadyExists{Key: fmt.Sprintf("%s/%s", req.Namespace, req.Name)}
			}
		}
		return &store.ErrInternal{Message: err.Error()}
	}

	return nil
}

func (s *StoreV2) Get(req storev2.ResourceRequest) (storev2.Wrapper, error) {
	if !req.UsePostgres {
		return s.etcdStoreV2.Get(req)
	}
	query, ok := s.lookupQuery(req, getQ)
	if !ok {
		return nil, errNoQuery(req.StoreName)
	}
	row := s.db.QueryRow(req.Context, query, req.Namespace, req.Name)
	// TODO(jk): make a generic wrapper type here or have the method take a
	// wrapper as an argument.
	var wrapper EntityWrapper
	if err := row.Scan(wrapper.SQLParams()...); err != nil {
		if err == pgx.ErrNoRows {
			return nil, &store.ErrNotFound{Key: fmt.Sprintf("%s.%s", req.Namespace, req.Name)}
		}
		return nil, &store.ErrInternal{Message: err.Error()}
	}
	return &wrapper, nil
}

func (s *StoreV2) GetMultiple(ctx context.Context, reqs []storev2.ResourceRequest) (map[storev2.ResourceRequest]storev2.Wrapper, error) {
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
		key := requestKey{
			Namespace: req.Namespace,
			Name:      req.Name,
		}
		keyedResourceRequests[key] = req
		namespacedResources[req.Namespace] = append(namespacedResources[req.Namespace], req.Name)
	}

	wrappers := map[storev2.ResourceRequest]storev2.Wrapper{}
	for namespace, resourceNames := range namespacedResources {
		// TODO(jk): this is opinionated and needs to change if we want to
		// support anything other than entity states
		rows, err := tx.Query(ctx, getEntityStatesQuery, namespace, resourceNames)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var wrapper EntityWrapper
			if err := rows.Scan(wrapper.SQLParams()...); err != nil {
				if err == pgx.ErrNoRows {
					continue
				}
				return nil, &store.ErrInternal{Message: err.Error()}
			}

			key := requestKey{
				Namespace: namespace,
				Name:      wrapper.Name,
			}
			req, ok := keyedResourceRequests[key]
			if !ok {
				panic("keyedResourceRequests key does not exist")
			}

			wrappers[req] = &wrapper
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error committing transaction for GetMultiple()")
	}

	return wrappers, nil
}

func (s *StoreV2) Delete(req storev2.ResourceRequest) error {
	if !req.UsePostgres {
		return s.etcdStoreV2.Delete(req)
	}
	query, ok := s.lookupQuery(req, deleteQ)
	if !ok {
		return errNoQuery(req.StoreName)
	}
	result, err := s.db.Exec(req.Context, query, req.Namespace, req.Name)
	if err != nil {
		return &store.ErrInternal{Message: err.Error()}
	}
	affected := result.RowsAffected()
	if affected < 1 {
		return &store.ErrNotFound{Key: fmt.Sprintf("%s.%s", req.Namespace, req.Name)}
	}
	return nil
}

type WrapList []storev2.Wrapper

func (w WrapList) Unwrap() ([]corev3.Resource, error) {
	result := make([]corev3.Resource, len(w))
	return result, w.UnwrapInto(&result)
}

func (w WrapList) UnwrapInto(dest interface{}) error {
	switch list := dest.(type) {
	case *[]corev3.EntityState:
		return w.unwrapIntoEntityList(list)
	case *[]*corev3.EntityState:
		return w.unwrapIntoEntityPointerList(list)
	case *[]corev3.Resource:
		return w.unwrapIntoResourceList(list)
	default:
		return fmt.Errorf("can't unwrap list into %T", dest)
	}
}

func (w WrapList) unwrapIntoEntityList(list *[]corev3.EntityState) error {
	if len(*list) != len(w) {
		*list = make([]corev3.EntityState, len(w))
	}
	for i, entity := range w {
		ptr := &((*list)[i])
		if err := entity.UnwrapInto(ptr); err != nil {
			return err
		}
	}
	return nil
}

func (w WrapList) unwrapIntoEntityPointerList(list *[]*corev3.EntityState) error {
	if len(*list) != len(w) {
		*list = make([]*corev3.EntityState, len(w))
	}
	for i, entity := range w {
		ptr := (*list)[i]
		if ptr == nil {
			ptr = new(corev3.EntityState)
			(*list)[i] = ptr
		}
		if err := entity.UnwrapInto(ptr); err != nil {
			return err
		}
	}
	return nil
}

// todo: make this work generically
func (w WrapList) unwrapIntoResourceList(list *[]corev3.Resource) error {
	if len(*list) != len(w) {
		*list = make([]corev3.Resource, len(w))
	}
	for i, resource := range w {
		ptr := (*list)[i]
		if ptr == nil {
			ptr = new(corev3.EntityState)
			(*list)[i] = ptr
		}
		if err := resource.UnwrapInto(ptr); err != nil {
			return err
		}
	}
	return nil
}

func (w WrapList) Len() int {
	return len(w)
}

func (s *StoreV2) List(req storev2.ResourceRequest, pred *store.SelectionPredicate) (list storev2.WrapList, err error) {
	if !req.UsePostgres {
		return s.etcdStoreV2.List(req, pred)
	}
	query, ok := s.lookupQuery(req, listQ)
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
	rows, rerr := s.db.Query(req.Context, query, namespace, limit, offset)
	if rerr != nil {
		return nil, &store.ErrInternal{Message: rerr.Error()}
	}
	defer rows.Close()
	wrapList := WrapList{}
	var rowCount int64
	for rows.Next() {
		rowCount++
		var wrapper EntityWrapper
		if err := rows.Scan(wrapper.SQLParams()...); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		if err := rows.Err(); err != nil {
			return nil, &store.ErrInternal{Message: err.Error()}
		}
		wrapList = append(wrapList, &wrapper)
	}
	if pred != nil {
		if rowCount < pred.Limit {
			pred.Continue = ""
		}
	}
	return wrapList, nil
}

func (s *StoreV2) Exists(req storev2.ResourceRequest) (bool, error) {
	if !req.UsePostgres {
		return s.etcdStoreV2.Exists(req)
	}
	query, ok := s.lookupQuery(req, existsQ)
	if !ok {
		return false, errNoQuery(req.StoreName)
	}
	row := s.db.QueryRow(req.Context, query, req.Namespace, req.Name)
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

func (s *StoreV2) Patch(req storev2.ResourceRequest, w storev2.Wrapper, patcher patch.Patcher, conditions *store.ETagCondition) (err error) {
	if !req.UsePostgres {
		return s.etcdStoreV2.Patch(req, w, patcher, conditions)
	}

	if err := req.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	query, ok := s.lookupQuery(req, getQ)
	if !ok {
		return errNoQuery(req.StoreName)
	}

	var originalWrapper EntityWrapper

	tx, txerr := s.db.Begin(req.Context)
	if err != nil {
		return &store.ErrInternal{Message: txerr.Error()}
	}
	defer func() {
		if err == nil {
			err = tx.Commit(req.Context)
			return
		}
		if txerr := tx.Rollback(context.Background()); txerr != nil {
			err = txerr
		}
	}()

	row := tx.QueryRow(req.Context, query, req.Namespace, req.Name)
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

	pwrap, ok := wrapped.(wrapperParams)
	if !ok {
		return &store.ErrNotValid{Err: fmt.Errorf("bad wrapper type for postgres: %T", wrapped)}
	}

	query, ok = s.lookupQuery(req, createOrUpdateQ)
	if !ok {
		return errNoQuery(req.StoreName)
	}

	params := pwrap.SQLParams()

	if _, err := tx.Exec(req.Context, query, params...); err != nil {
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
