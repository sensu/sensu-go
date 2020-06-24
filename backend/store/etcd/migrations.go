package etcd

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// A migration is a function that receives a context and an etcd client, and
// returns an error. Migrations should be able to recover from a partial run.
type Migration func(ctx context.Context, client *clientv3.Client) error

// Migrations is the list of migrations. It must never be re-ordered. It can
// only be appended to.
var Migrations = []Migration{
	Base,
	MigrateV2EntityToV3,
}

// Base is the base version of the database. It is never executed.
func Base(ctx context.Context, client *clientv3.Client) error {
	return nil
}

// In Sensu 6.0, we migrate v2 entities to v3.
func MigrateV2EntityToV3(ctx context.Context, client *clientv3.Client) error {
	s := NewStore(client, "")
	responses := readPagedV2Entities(ctx, client)
	for response := range responses {
		if response.Err != nil {
			return response.Err
		}
		ctx := store.NamespaceContext(ctx, response.Entity.Namespace)
		if err := s.UpdateEntity(ctx, response.Entity); err != nil {
			return err
		}
		if err := deleteV2Entity(ctx, client, response.Entity); err != nil {
			return err
		}
	}
	return nil
}

func deleteV2Entity(ctx context.Context, client *clientv3.Client, entity *corev2.Entity) error {
	err := Delete(ctx, client, getEntityPath(entity))
	if _, ok := err.(*store.ErrNotFound); ok {
		err = nil
	}
	return err
}

type entityOrError struct {
	Entity *corev2.Entity
	Err    error
}

func readPagedV2Entities(ctx context.Context, client *clientv3.Client) <-chan entityOrError {
	const pageSize = 100
	result := make(chan entityOrError, pageSize)
	go func() {
		pred := &store.SelectionPredicate{Limit: 100}
		for {
			entities := []*corev2.Entity{}
			err := List(ctx, client, GetEntitiesPath, &entities, pred)
			if err != nil {
				result <- entityOrError{Err: err}
				close(result)
				return
			}
			for _, entity := range entities {
				result <- entityOrError{Entity: entity}
			}
			if pred.Continue == "" {
				close(result)
				return
			}
		}
	}()
	return result
}

// MigrateDB brings a database up to the most current version.
func MigrateDB(ctx context.Context, client *clientv3.Client, migrations []Migration) error {
	ver, err := GetDatabaseVersion(ctx, client)
	if err != nil {
		return fmt.Errorf("can't migrate database: %s", err)
	}
	if len(migrations) == ver+1 {
		logger.WithField("database_version", ver).Info("database already up to date")
		return nil
	}
	logger.Warn("migrating etcd database to a new version")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go notifyUpgradeLoop(ctx)
	for i := ver + 1; i < len(migrations); i++ {
		if err := doMigration(ctx, client, i, migrations[i]); err != nil {
			logger.WithField("database_version", i).Error("error upgrading database")
			return err
		}
		logger.WithField("database_version", i).Info("successfully upgraded database")
	}
	return nil
}

func notifyUpgradeLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			logger.Warn("upgrading database")
		case <-ctx.Done():
			return
		}
	}
}

func versionCmp(version int) clientv3.Cmp {
	return clientv3.Compare(
		clientv3.Value(DatabaseVersionKey),
		"=",
		fmt.Sprintf("%d", version))
}

func versionOp(version int) clientv3.Op {
	return clientv3.OpPut(DatabaseVersionKey, fmt.Sprintf("%d", version))
}

func doMigration(ctx context.Context, client *clientv3.Client, version int, do func(context.Context, *clientv3.Client) error) error {
	if err := do(ctx, client); err != nil {
		return err
	}

	return SetDatabaseVersion(ctx, client, version)
}
