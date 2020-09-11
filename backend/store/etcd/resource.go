package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/textproto"
	"strings"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/patch"
)

// CreateResource creates the given resource only if it does not already exist
func (s *Store) CreateResource(ctx context.Context, resource corev2.Resource) error {
	if err := resource.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	key := store.KeyFromResource(resource)
	namespace := resource.GetObjectMeta().Namespace

	msg, ok := resource.(proto.Message)
	if !ok {
		return &store.ErrEncode{Key: key, Err: fmt.Errorf("%T is not proto.Message", resource)}
	}

	return Create(ctx, s.client, key, namespace, msg)
}

// CreateOrUpdateResource creates or updates the given resource regardless of
// whether it already exists or not
func (s *Store) CreateOrUpdateResource(ctx context.Context, resource corev2.Resource) error {
	if err := resource.Validate(); err != nil {
		return &store.ErrNotValid{Err: err}
	}

	key := store.KeyFromResource(resource)
	namespace := resource.GetObjectMeta().Namespace
	return CreateOrUpdate(ctx, s.client, key, namespace, resource)
}

// DeleteResource deletes the resource using the given resource prefix and name
func (s *Store) DeleteResource(ctx context.Context, resourcePrefix, name string) error {
	key := store.KeyFromArgs(ctx, resourcePrefix, name)
	return Delete(ctx, s.client, key)
}

// GetResource retrieves a resource with the given name and stores it into the
// resource pointer
func (s *Store) GetResource(ctx context.Context, name string, resource corev2.Resource) error {
	key := store.KeyFromArgs(ctx, resource.StorePrefix(), name)
	return Get(ctx, s.client, key, resource)
}

// ListResources retrieves all resources for the resourcePrefix type and stores
// them into the resources pointer
func (s *Store) ListResources(ctx context.Context, resourcePrefix string, resources interface{}, pred *store.SelectionPredicate) error {
	keyBuilderFunc := func(ctx context.Context, name string) string {
		return store.NewKeyBuilder(resourcePrefix).WithContext(ctx).Build("")
	}

	return List(ctx, s.client, keyBuilderFunc, resources, pred)
}

func (s *Store) PatchResource(ctx context.Context, resource corev2.Resource, name string, patcher patch.Patcher, conditions *store.ETagCondition) error {
	key := store.KeyFromArgs(ctx, resource.StorePrefix(), name)

	// Get the stored resource along with the etcd response so we can use the
	// revision later to ensure the resource wasn't modified in the mean time
	resp, err := GetResponse(ctx, s.client, key, resource)
	if err != nil {
		return err
	}
	value := resp.Kvs[0].Value

	// Determine the etag for the stored value
	etag, err := ETag(resource)
	if err != nil {
		return err
	}

	if conditions != nil {
		if !checkIfMatch(conditions.IfMatch, etag) {
			return &store.ErrModified{Key: key}
		}
		if !checkIfNoneMatch(conditions.IfNoneMatch, etag) {
			return &store.ErrModified{Key: key}
		}
	}

	// Encode the stored resource
	original, err := json.Marshal(resource)
	if err != nil {
		return err
	}

	// Apply the patch to our original document (stored resource)
	patchedResource, err := patcher.Patch(original)
	if err != nil {
		return err
	}

	// Decode the resulting document into provided resource
	if err := json.Unmarshal(patchedResource, &resource); err != nil {
		return err
	}

	return UpdateWithValue(ctx, s.client, key, resource, value)
}

// checkIfMatch determines if any of the etag provided in the If-Match header
// match the stored etag. This function was largely inspired by the net/http
// package
func checkIfMatch(header string, etag string) bool {
	if header == "" {
		return true
	}

	for {
		header = textproto.TrimString(header)
		if len(header) == 0 {
			break
		}
		if header[0] == ',' {
			header = header[1:]
			continue
		}
		if header[0] == '*' {
			return true
		}
		scannedEtag, remainingHeader := scanETag(header)
		if scannedEtag == etag && scannedEtag != "" && scannedEtag[0] == '"' {
			return true
		}
		header = remainingHeader
	}

	return false
}

// checkIfNoneMatch determines if none of the etag provided in the If-Match
// header match the stored etag. This function was largely inspired by the
// net/http package
func checkIfNoneMatch(header string, etag string) bool {
	if header == "" {
		return true
	}

	for {
		header = textproto.TrimString(header)
		if len(header) == 0 {
			break
		}
		if header[0] == ',' {
			header = header[1:]
			continue
		}
		if header[0] == '*' {
			return false
		}
		scannedEtag, remainingHeader := scanETag(header)
		if strings.TrimPrefix(scannedEtag, "W/") == strings.TrimPrefix(etag, "W/") {
			return false
		}
		header = remainingHeader
	}

	return true
}

func scanETag(header string) (string, string) {
	fmt.Printf("scanETag header = %s\n", header)
	header = textproto.TrimString(header)
	start := 0
	if strings.HasPrefix(header, "W/") {
		start = 2
	}

	if len(header[start:]) < 2 || header[start] != '"' {
		return "", ""
	}

	// ETag is either W/"text" or "text".
	// See RFC 7232 2.3.
	for i := start + 1; i < len(header); i++ {
		c := header[i]
		switch {
		// Character values allowed in ETags.
		case c == 0x21 || c >= 0x23 && c <= 0x7E || c >= 0x80:
		case c == '"':
			fmt.Printf("1: %s\n2: %s\n", string(header[:i+1]), header[i+1:])
			return string(header[:i+1]), header[i+1:]
		default:
			break
		}
	}

	return "", ""
}
