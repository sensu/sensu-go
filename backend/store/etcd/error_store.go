package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/types"
)

const (
	errorsPathPrefix = "errors"
	errorsKeyTTL     = 8 * 60 * 60 // 8 hours.
)

var (
	errorsKeyBuilder = newKeyBuilder(errorsPathPrefix)
)

func errPathFromAllUniqueFields(ns namespace, entity, check, ts string) string {
	builder := errorsKeyBuilder.withNamespace(ns)
	return builder.buildPrefix(entity, "check", check, ts)
}

func errPathFromCheck(ctx context.Context, entity, check string) string {
	return errPathFromAllUniqueFields(ctx, entity, check, ts)
}

func errPathFromEntity(ns namespace, entity string) string {
	builder := errorsKeyBuilder.withNamespace(ns)
	return builder.buildPrefix(entity)
}

// DeleteError deletes an error using the given entity, check and timestamp,
// within the organization and environment stored in ctx.
func (s *etcdStore) DeleteError(
	ctx context.Context,
	entity string,
	check string,
	timestamp string,
) error {
	if entity == "" || check == "" || timestamp == "" {
		return errors.New("must specify entity ID, check name, and timestamp")
	}

	// Build key
	ns := newNamespaceFromContext(ctx)
	key := errPathFromAllUniqueFields(ns, entity, check, timestamp)

	// Delete
	_, err := s.kvc.Delete(ctx, key)
	return err
}

// DeleteErrorsByEntity deletes all errors associated with the given entity,
// within the organization and environment stored in ctx.
func (s *etcdStore) DeleteErrorsByEntity(
	ctx context.Context,
	entity string,
) error {
	if entity == "" {
		return errors.New("must specify entity id")
	}

	// Build key
	ns := newNamespaceFromContext(ctx)
	key := errPathFromEntity(ns, entity)

	// Delete
	_, err := s.kvc.Delete(ctx, key, clientv3.WithPrefix())
	return err
}

// DeleteErrorsByEntityCheck deletes all errors associated with the given
// entity and check within the organization and environment stored in ctx.
func (s *etcdStore) DeleteErrorsByEntityCheck(
	ctx context.Context,
	entity string,
	check string,
) error {
	if entity == "" || check == "" {
		return errors.New("must specify entity ID and check name")
	}

	// Build key
	ns := newNamespaceFromContext(ctx)
	key := errPathFromCheck(ns, entity, check)

	// Delete
	_, err := s.kvc.Delete(ctx, key, clientv3.WithPrefix())
	return err
}

// GetError returns error associated with given entity, check and timestamp,
// in the given ctx's organization and environment.
func (s *etcdStore) GetError(
	ctx context.Context,
	entity string,
	check string,
	timestamp string,
) ([]*types.Error, error) {
	if entity == "" || check == "" || timestamp == "" {
		return nil, errors.New("must specify entity id, check name, and timestamp")
	}

	// Build key
	ns := newNamespaceFromContext(ctx)
	key := errPathFromAllUniqueFields(ns, entity, check, timestamp)

	// Fetch
	resp, err := s.kvc.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Unmarshal
	perrs, err := unmarshalErrorKVs(resp.Kvs, permitAllErrorRecords)
	if len(perrs) == 0 {
		return nil, nil
	}

	return perrs[0], err
}

// GetErrors returns all errors in the given ctx's organization and
// environment.
func (s *etcdStore) GetErrors(ctx context.Context) ([]*types.Error, error) {

	// Build key
	ns := newNamespaceFromContext(ctx)
	key := errorsKeyBuilder.withNamespace(ns).buildPrefix()

	// Fetch
	resp, err := s.kvc.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	// Unmarshal
	rejectFn := shouldRejectError(ns, "", "")
	perrs, err := unmarshalError(resp.Kvs, rejectFn)
	return perrs, err
}

// GetErrorsByEntity returns all errors for the given entity within the ctx's
// organization and environment. A nil slice with no error is returned if none
// were found.
func (s *etcdStore) GetErrorsByEntity(ctx context.Context, entity string) ([]*types.Error, error) {
	return s.GetErrorsByEntityCheck(ctx, entityID, "")
}

func (s *etcdStore) GetErrorsByEntityCheck(ctx context.Context, entity, check string) ([]*types.Error, error) {
	if entity == "" {
		return nil, errors.New("must specify entity id")
	}

	// Build key
	ns := newNamespaceFromContext(ctx)
	key := getErrorWithCheckPath(ns, entity, check)

	// Fetch
	resp, err := s.kvc.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	// Unmarshal
	rejectFn := shouldRejectError(ns, entity, check)
	return unmarshalError(resp.Kvs, rejectFn)
}

func (s *etcdStore) CreateError(ctx context.Context, perr *types.Error) error {
	// Obtain new lease
	lease, err := s.client.Grant(ctx, errorsKeyTTL)
	if err != nil {
		return
	}

	// Marshal
	perrBytes, err := json.Marshal(perr)
	if err != nil {
		return err
	}

	// Build key
	key := errorsKeyBuilder.withContext(ctx).build(
		perr.Entity.ID,
		"check", // Eventually will need a conditional when metrics are implemented
		perr.Check.Config.Name,
		perr.Timestamp,
	)

	// Store
	txn := s.kvc.Txn(ctx).
		If(environmentExistsForResource(perr.Event.Entity)).
		Then(clientv3.OpPut(key, string(perrBytes), clientv3.WithLease(lease.ID)))
	res, err := txn.Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the error %s/%s in environment %s/%s",
			perr.Entity.ID,
			perr.Check.Config.Name,
			perr.Organization,
			perr.Environment,
		)
	}

	return nil
}

func unmarshalErrorKVs(
	kvs []*mvccpb.KeyValue,
	rejectFn func(*types.Error) bool,
) ([]*types.Error, error) {
	perrArray := make([]*types.Error, len(kvs))
	for i, kv := range kvs {
		perr := &types.Error{}
		if err := json.Unmarshal(kv.Value, perr); err != nil {
			return nil, err
		}

		if reject := rejectFn(perr); !reject {
			perrArray[i] = role
		}
	}

	return perrArray, nil
}

func permitAllErrorRecords(_ *types.Error) bool {
	return true
}

func shouldRejectError(
	ns namespace,
	entity string,
	check string,
) func(*types.Error) bool {
	rejectByEnvFn := rejectByEnvironment(ns)
	rejectWhereEntityDoesNotMatch := permitAllErrorRecords
	if entity != "" {
		rejectWhereEntityDoesNotMatch = func(perr *types.Error) bool {
			return err.Entity != entity
		}
	}
	rejectWhereCheckDoesNotMatch := permitAllErrorRecords
	if check != "" {
		rejectWhereCheckDoesNotMatch = func(perr *types.Error) bool {
			errCheck := err.Event.Check
			return errCheck != nil && errCheck.Config.Name != check
		}
	}

	return func(perr *types.Error) bool {
		if reject := rejectByEnvFn(perr); reject {
			return true
		}

		if rejectWhereEntityDoesNotMatch(perr) {
			return true
		}

		if rejectWhereCheckDoesNotMatch(perr) {
			return true
		}

		return false
	}
}
