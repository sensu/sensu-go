package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	v2 "github.com/sensu/sensu-go/api/core/v2"
	v3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/types"
)

const pgUniqueViolationCode = "23505"

type ConfigStore struct {
	db *pgxpool.Pool
}

type DBResource struct {
	id          int64
	apiVersion  string
	apiType     string
	namespace   string
	name        string
	labels      string
	annotations []byte
	resource    string
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   time.Time
}

func NewConfigStore(db *pgxpool.Pool) *ConfigStore {
	return &ConfigStore{
		db: db,
	}
}

func (s *ConfigStore) CreateOrUpdate(request storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	if err := request.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	typeMeta, meta, jsonLabels, bytesAnnotations, jsonResource, err := extractResourceData(wrapper)
	if err != nil {
		return err
	}

	args := []interface{}{typeMeta.APIVersion, typeMeta.Type, meta.Namespace, meta.Name, jsonLabels, bytesAnnotations, jsonResource}

	tags, err := s.db.Exec(request.Context, CreateOrUpdateConfigQuery, args...)
	if err != nil {
		return err
	}
	if tags.RowsAffected() == 0 {
		return errors.New("no rows inserted")
	}

	return nil
}

func (s *ConfigStore) UpdateIfExists(request storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	if err := request.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	typeMeta, meta, jsonLabels, bytesAnnotations, jsonResource, err := extractResourceData(wrapper)
	if err != nil {
		return err
	}

	args := []interface{}{jsonLabels, bytesAnnotations, jsonResource, typeMeta.APIVersion, typeMeta.Type, meta.Namespace, meta.Name}

	tags, err := s.db.Exec(request.Context, UpdateIfExistsConfigQuery, args...)
	if err != nil {
		return err
	}
	if tags.RowsAffected() == 0 {
		return &store.ErrNotFound{Key: fmt.Sprintf("%s.%s/%s/%s", typeMeta.APIVersion, typeMeta.Type, request.Namespace, request.Name)}
	}

	return nil
}

func (s *ConfigStore) CreateIfNotExists(request storev2.ResourceRequest, wrapper storev2.Wrapper) error {
	if err := request.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	typeMeta, meta, jsonLabels, bytesAnnotations, jsonResource, err := extractResourceData(wrapper)
	if err != nil {
		return err
	}

	args := []interface{}{typeMeta.APIVersion, typeMeta.Type, meta.Namespace, meta.Name, jsonLabels, bytesAnnotations, jsonResource}

	tags, err := s.db.Exec(request.Context, CreateConfigIfNotExistsQuery, args...)
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

func (s *ConfigStore) Get(request storev2.ResourceRequest) (storev2.Wrapper, error) {
	if err := request.Validate(); err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}

	typeMeta, err := types.ResolveTypeMeta(request.StoreName)
	if err != nil {
		return nil, err
	}

	args := []interface{}{typeMeta.APIVersion, typeMeta.Type, request.Namespace, request.Name}

	row := s.db.QueryRow(request.Context, GetConfigQuery, args...)
	result := DBResource{}
	err = row.Scan(&result.id, &result.labels, &result.annotations, &result.resource, &result.createdAt, &result.updatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &store.ErrNotFound{Key: fmt.Sprintf("%s.%s/%s/%s", typeMeta.APIVersion, typeMeta.Type, request.Namespace, request.Name)}
		}
		return nil, err
	}

	return &wrap.Wrapper{
		TypeMeta:    &typeMeta,
		Encoding:    wrap.Encoding_json,
		Compression: 0,
		Value:       []byte(result.resource),
	}, nil
}

func (s *ConfigStore) Delete(request storev2.ResourceRequest) error {
	if err := request.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	typeMeta, err := types.ResolveTypeMeta(request.StoreName)
	if err != nil {
		return err
	}

	args := []interface{}{typeMeta.APIVersion, typeMeta.Type, request.Namespace, request.Name}

	cmdTag, err := s.db.Exec(request.Context, DeleteConfigQuery, args...)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return &store.ErrNotFound{Key: fmt.Sprintf("%s.%s/%s/%s", typeMeta.APIVersion, typeMeta.Type, request.Namespace, request.Name)}
	}

	return nil
}

func (s *ConfigStore) List(request storev2.ResourceRequest, predicate *store.SelectionPredicate) (storev2.WrapList, error) {
	if err := request.Validate(); err != nil {
		return nil, &store.ErrNotValid{Err: err}
	}

	typeMeta, err := types.ResolveTypeMeta(request.StoreName)
	if err != nil {
		return nil, err
	}

	//selector := selector.SelectorFromContext(request.Context)

	tmpl, err := template.New("listResourceQuery").Parse(ListConfigQueryTmpl)
	if err != nil {
		return nil, err
	}
	var queryBuilder strings.Builder
	if predicate == nil {
		predicate = &store.SelectionPredicate{}
	}
	if err := tmpl.Execute(&queryBuilder, predicate); err != nil {
		return nil, err
	}

	args := []interface{}{typeMeta.APIVersion, typeMeta.Type, request.Namespace}

	query := queryBuilder.String()
	rows, err := s.db.Query(request.Context, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wrapList := make(wrap.List, 0)
	for rows.Next() {
		var dbResource DBResource
		err := rows.Scan(&dbResource.id, &dbResource.labels, &dbResource.annotations, &dbResource.resource, &dbResource.createdAt, &dbResource.updatedAt)
		if err != nil {
			return nil, err
		}

		wrapped := wrap.Wrapper{
			TypeMeta:    &typeMeta,
			Encoding:    wrap.Encoding_json,
			Compression: wrap.Compression_none,
			Value:       []byte(dbResource.resource),
		}
		wrapList = append(wrapList, &wrapped)
	}

	return wrapList, nil
}

func (s *ConfigStore) Exists(request storev2.ResourceRequest) (bool, error) {
	if err := request.Validate(); err != nil {
		return false, &store.ErrNotValid{Err: err}
	}

	typeMeta, err := types.ResolveTypeMeta(request.StoreName)
	if err != nil {
		return false, err
	}

	args := []interface{}{typeMeta.APIVersion, typeMeta.Type, request.Namespace, request.Name}

	row := s.db.QueryRow(request.Context, ExistsConfigQuery, args...)

	var count int
	err = row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *ConfigStore) Patch(request storev2.ResourceRequest, wrapper storev2.Wrapper, patcher patch.Patcher, condition *store.ETagCondition) error {
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

	if err := json.Unmarshal(patchedResource, resource); err != nil {
		return err
	}

	if err := resource.Validate(); err != nil {
		return err
	}

	// Special case for entities; we need to make sure we keep the per-entity
	// subscription
	if e, ok := resource.(*v3.EntityConfig); ok {
		e.Subscriptions = v2.AddEntitySubscription(e.Metadata.Name, e.Subscriptions)
	}

	// Re-wrap the resource
	wrappedPatch, err := wrap.Resource(resource)
	if err != nil {
		return &store.ErrEncode{Key: key, Err: err}
	}
	*w = *wrappedPatch

	return s.UpdateIfExists(request, w)
}

func labelsToJSON(labels map[string]string) (string, error) {
	if labels == nil || len(labels) == 0 {
		return "{}", nil
	}

	bytes, err := json.Marshal(labels)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func annotationsToBytes(annotations map[string]string) ([]byte, error) {
	if annotations == nil || len(annotations) == 0 {
		return []byte("{}"), nil
	}

	return json.Marshal(annotations)
}

func extractResourceData(wrapper storev2.Wrapper) (typeMeta v2.TypeMeta, meta *v2.ObjectMeta, jsonLabels string,
	bytesAnnotations, jsonResource []byte, err error) {
	res, err := wrapper.Unwrap()
	//res, err := wrapper.UnwrapRaw()
	if err != nil {
		return
	}

	if resV3, ok := res.(v3.Resource); ok {
		meta = resV3.GetMetadata()
	} else if resV2, ok := res.(v2.Resource); ok {
		objectMeta := resV2.GetObjectMeta()
		meta = &objectMeta
	}

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

	resProxy := v3.V2ResourceProxy{Resource: res}
	typeMeta = resProxy.GetTypeMeta()

	return
}

// StoreKey converts a ResourceRequest into a key that uniquely identifies a
// singular resource, or collection of resources, in a namespace.
func storeKey(req storev2.ResourceRequest) string {
	return store.NewKeyBuilder(req.StoreName).WithNamespace(req.Namespace).Build(req.Name)
}
