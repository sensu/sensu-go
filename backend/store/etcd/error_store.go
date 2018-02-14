package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

const (
	errorsPathPrefix = "errors"
	errorsKeyTTL     = 8 * 60 * 60 // 8 hours.
)

var (
	errorsKeyBuilder = store.NewKeyBuilder(errorsPathPrefix)
)

func errPathFromAllUniqueFields(ns store.Namespace, entity, check, ts string) string {
	builder := errorsKeyBuilder.WithNamespace(ns)
	return builder.Build(entity, "check", check, ts)
}

func errPathFromCheck(ns store.Namespace, entity, check string) string {
	return errPathFromAllUniqueFields(ns, entity, check, "")
}

func errPathFromEntity(ns store.Namespace, entity string) string {
	builder := errorsKeyBuilder.WithNamespace(ns)
	return builder.BuildPrefix(entity)
}

// DeleteError deletes an error using the given entity, check and timestamp,
// within the organization and environment stored in ctx.
func (s *Store) DeleteError(
	ctx context.Context,
	entity string,
	check string,
	timestamp string,
) error {
	if entity == "" || check == "" || timestamp == "" {
		return errors.New("must specify entity ID, check name, and timestamp")
	}

	// Build key
	ns := store.NewNamespaceFromContext(ctx)
	key := errPathFromAllUniqueFields(ns, entity, check, timestamp)

	// Delete
	_, err := s.kvc.Delete(ctx, key)
	return err
}

// DeleteErrorsByEntity deletes all errors associated with the given entity,
// within the organization and environment stored in ctx.
func (s *Store) DeleteErrorsByEntity(
	ctx context.Context,
	entity string,
) error {
	if entity == "" {
		return errors.New("must specify entity id")
	}

	// Build key
	ns := store.NewNamespaceFromContext(ctx)
	key := errPathFromEntity(ns, entity)

	// Delete
	_, err := s.kvc.Delete(ctx, key, clientv3.WithPrefix())
	return err
}

// DeleteErrorsByEntityCheck deletes all errors associated with the given
// entity and check within the organization and environment stored in ctx.
func (s *Store) DeleteErrorsByEntityCheck(
	ctx context.Context,
	entity string,
	check string,
) error {
	if entity == "" || check == "" {
		return errors.New("must specify entity ID and check name")
	}

	// Build key
	ns := store.NewNamespaceFromContext(ctx)
	key := errPathFromCheck(ns, entity, check)

	// Delete
	_, err := s.kvc.Delete(ctx, key, clientv3.WithPrefix())
	return err
}

// GetError returns error associated with given entity, check and timestamp,
// in the given ctx's organization and environment.
func (s *Store) GetError(
	ctx context.Context,
	entity string,
	check string,
	timestamp string,
) (*types.Error, error) {
	// Build key
	ns := store.NewNamespaceFromContext(ctx)
	key := errPathFromAllUniqueFields(ns, entity, check, timestamp)

	// Validate arguments
	if entity == "" || check == "" || timestamp == "" {
		return nil, errors.New("must specify entity id, check name, and timestamp")
	} else if ns.Wildcard() {
		return nil, errors.New("may not use wildcard to search for record")
	}

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
func (s *Store) GetErrors(ctx context.Context) ([]*types.Error, error) {

	// Build key
	ns := store.NewNamespaceFromContext(ctx)
	key := errorsKeyBuilder.WithNamespace(ns).BuildPrefix()

	// Fetch
	resp, err := s.kvc.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	// Unmarshal
	rejectFn := shouldRejectError(ns, "", "")
	perrs, err := unmarshalErrorKVs(resp.Kvs, rejectFn)
	return perrs, err
}

// GetErrorsByEntity returns all errors for the given entity within the ctx's
// organization and environment. A nil slice with no error is returned if none
// were found.
func (s *Store) GetErrorsByEntity(ctx context.Context, entity string) ([]*types.Error, error) {
	return s.GetErrorsByEntityCheck(ctx, entity, "")
}

// GetErrorsByEntityCheck returns an error using the given entity and check,
// within the organization and environment stored in ctx. The resulting error
// is nil if none was found.
func (s *Store) GetErrorsByEntityCheck(ctx context.Context, entity, check string) ([]*types.Error, error) {
	if entity == "" {
		return nil, errors.New("must specify entity id")
	}

	// Build key
	ns := store.NewNamespaceFromContext(ctx)
	key := errPathFromCheck(ns, entity, check)

	// Fetch
	resp, err := s.kvc.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	// Unmarshal
	rejectFn := shouldRejectError(ns, entity, check)
	return unmarshalErrorKVs(resp.Kvs, rejectFn)
}

// CreateError creates or updates a given error.
func (s *Store) CreateError(ctx context.Context, perr *types.Error) error {
	// Obtain new lease
	lease, err := s.client.Grant(ctx, errorsKeyTTL)
	if err != nil {
		return err
	}

	// Marshal
	perrBytes, err := json.Marshal(perr)
	if err != nil {
		return err
	}

	// Build key
	key := errorsKeyBuilder.WithContext(ctx).Build(
		perr.Event.Entity.ID,
		"check", // Eventually will need a conditional when metrics are implemented
		perr.Event.Check.Name,
		string(perr.Timestamp),
	)

	// Configure transaction
	txn := s.kvc.Txn(ctx).
		If(environmentExistsForResource(perr.Event.Entity)).
		Then(clientv3.OpPut(key, string(perrBytes), clientv3.WithLease(lease.ID)))

	// Store
	res, err := txn.Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf(
			"could not create the error %s/%s in environment %s/%s",
			perr.Event.Entity.ID,
			perr.Event.Check.Name,
			perr.GetOrganization(),
			perr.GetEnvironment(),
		)
	}

	return nil
}

type rejectErrorFn func(*types.Error) bool

func unmarshalErrorKVs(
	kvs []*mvccpb.KeyValue,
	rejectFn rejectErrorFn,
) ([]*types.Error, error) {
	perrs := make([]*types.Error, 0, len(kvs))
	for _, kv := range kvs {
		perr := &types.Error{}
		if err := json.Unmarshal(kv.Value, perr); err != nil {
			return nil, err
		}

		if shouldReject := rejectFn(perr); !shouldReject {
			perrs = append(perrs, perr)
		}
	}

	return perrs, nil
}

func permitAllErrorRecords(_ *types.Error) bool {
	return false
}

// shouldRejectError configures a predicate to be used when rejecting error
// entries. If entity argument is an empty string it will not reject based on
// the error's entity; same rule applies for check argument.
func shouldRejectError(ns store.Namespace, entity string, check string) rejectErrorFn {
	rejectWhereEnvDoesNotMatch := rejectByEnvironment(ns)
	rejectWhereEntityDoesNotMatch := permitAllErrorRecords
	rejectWhereCheckDoesNotMatch := permitAllErrorRecords

	if entity != "" {
		rejectWhereEntityDoesNotMatch = func(perr *types.Error) bool {
			return perr.Event.Entity.ID != entity
		}
	}

	if check != "" {
		rejectWhereCheckDoesNotMatch = func(perr *types.Error) bool {
			errCheck := perr.Event.Check
			return errCheck != nil && errCheck.Name != check
		}
	}

	return func(perr *types.Error) bool {
		if rejectWhereEnvDoesNotMatch(perr) {
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
