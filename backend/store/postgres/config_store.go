package postgres

import (
	"context"
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
	"github.com/sensu/sensu-go/backend/selector"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
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

type listTemplateValues struct {
	Limit       int64
	Offset      int64
	SelectorSQL string
}

func NewConfigStore(db *pgxpool.Pool) *ConfigStore {
	return &ConfigStore{
		db: db,
	}
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
	result := DBResource{}

	if err := row.Scan(&result.id, &result.labels, &result.annotations, &result.resource, &result.createdAt, &result.updatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, &store.ErrNotFound{Key: fmt.Sprintf("%s.%s/%s/%s", request.APIVersion, request.Type, request.Namespace, request.Name)}
		}
		return nil, err
	}

	return &wrap.Wrapper{
		TypeMeta:    &v2.TypeMeta{APIVersion: request.APIVersion, Type: request.Type},
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
		var dbResource DBResource
		err := rows.Scan(&dbResource.id, &dbResource.labels, &dbResource.annotations, &dbResource.resource, &dbResource.createdAt, &dbResource.updatedAt)
		if err != nil {
			return nil, err
		}

		wrapped := wrap.Wrapper{
			TypeMeta:    &v2.TypeMeta{APIVersion: request.APIVersion, Type: request.Type},
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

	return s.UpdateIfExists(ctx, request, w)
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

func getSelectorSQL(ctx context.Context) (string, []interface{}, error) {
	ctxSelector := selector.SelectorFromContext(ctx)

	if ctxSelector != nil && len(ctxSelector.Operations) > 0 {
		argCounter := argCounter{value: 3}
		builder := NewConfigSelectorSQLBuilder(ctxSelector)
		return builder.GetSelectorCond(&argCounter)
	}

	return "", nil, nil
}
