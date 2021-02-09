// +build !race,integration

package etcd

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

func testMigration(ctx context.Context, client *clientv3.Client) error {
	return nil
}

func testFailingMigration(ctx context.Context, client *clientv3.Client) error {
	return errors.New("error")
}

func TestMigrateDBBase(t *testing.T) {
	migrations := []Migration{Base}
	testWithEtcdClient(t, func(_ store.Store, client *clientv3.Client) {
		ctx := context.Background()
		err := MigrateDB(ctx, client, migrations)
		if err != nil {
			t.Fatal(err)
		}
		version, err := GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 0; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestMigrationDBBroken(t *testing.T) {
	migrations := []Migration{
		Base,
		testFailingMigration,
		testMigration,
	}
	testWithEtcdClient(t, func(_ store.Store, client *clientv3.Client) {
		ctx := context.Background()
		err := MigrateDB(ctx, client, migrations)
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		version, err := GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 0; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestMigrationDBLastBroken(t *testing.T) {
	migrations := []Migration{
		Base,
		testMigration,
		testFailingMigration,
	}
	testWithEtcdClient(t, func(_ store.Store, client *clientv3.Client) {
		ctx := context.Background()
		err := MigrateDB(ctx, client, migrations)
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		version, err := GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 1; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestMigrationDBGood(t *testing.T) {
	migrations := []Migration{
		Base,
		testMigration,
		testMigration,
	}
	testWithEtcdClient(t, func(_ store.Store, client *clientv3.Client) {
		ctx := context.Background()
		err := MigrateDB(ctx, client, migrations)
		if err != nil {
			t.Fatal(err)
		}
		version, err := GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 2; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestReapplyMigration(t *testing.T) {
	migrations := []Migration{
		Base,
		testMigration,
		testMigration,
	}
	testWithEtcdClient(t, func(_ store.Store, client *clientv3.Client) {
		ctx := context.Background()
		err := MigrateDB(ctx, client, migrations)
		if err != nil {
			t.Fatal(err)
		}
		version, err := GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 2; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
		if err := MigrateDB(ctx, client, migrations); err != nil {
			t.Fatal(err)
		}
		version, err = GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 2; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestUpgradeV2Entities(t *testing.T) {
	migrations := []Migration{
		Base,
		MigrateV2EntityToV3,
	}
	testWithEtcdClient(t, func(s store.Store, client *clientv3.Client) {
		ctx := store.NamespaceContext(context.Background(), "default")
		entities := make([]*corev2.Entity, 10)
		for i := range entities {
			entities[i] = corev2.FixtureEntity(fmt.Sprintf("%d", i))
		}
		for _, entity := range entities {
			key := getEntityPath(entity)
			if err := Create(ctx, client, key, "default", entity); err != nil {
				t.Fatal(err)
			}
		}
		if err := MigrateDB(ctx, client, migrations); err != nil {
			t.Fatal(err)
		}
		got, err := s.GetEntities(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		if want := entities; !equalEntities(got, want) {
			t.Errorf("bad entities: got %v, want %v", got, want)
		}
	})
}

func equalEntities(got, want []*corev2.Entity) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if !proto.Equal(got[i], want[i]) {
			return false
		}
	}
	return true
}

func TestMigrateEnterpriseDBBaseEnterprise(t *testing.T) {
	migrations := []Migration{Base}
	testWithEtcdClient(t, func(_ store.Store, client *clientv3.Client) {
		ctx := context.Background()
		err := MigrateEnterpriseDB(ctx, client, migrations)
		if err != nil {
			t.Fatal(err)
		}
		version, err := GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 0; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestMigrationDBBrokenEnterprise(t *testing.T) {
	migrations := []Migration{
		Base,
		testFailingMigration,
		testMigration,
	}
	testWithEtcdClient(t, func(_ store.Store, client *clientv3.Client) {
		ctx := context.Background()
		err := MigrateEnterpriseDB(ctx, client, migrations)
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		version, err := GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 0; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestMigrationDBLastBrokenEnterprise(t *testing.T) {
	migrations := []Migration{
		Base,
		testMigration,
		testFailingMigration,
	}
	testWithEtcdClient(t, func(_ store.Store, client *clientv3.Client) {
		ctx := context.Background()
		err := MigrateEnterpriseDB(ctx, client, migrations)
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		version, err := GetEnterpriseDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 1; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestMigrationDBGoodEnterprise(t *testing.T) {
	migrations := []Migration{
		Base,
		testMigration,
		testMigration,
	}
	testWithEtcdClient(t, func(_ store.Store, client *clientv3.Client) {
		ctx := context.Background()
		err := MigrateEnterpriseDB(ctx, client, migrations)
		if err != nil {
			t.Fatal(err)
		}
		version, err := GetEnterpriseDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 2; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestReapplyMigrationEnterprise(t *testing.T) {
	migrations := []Migration{
		Base,
		testMigration,
		testMigration,
	}
	testWithEtcdClient(t, func(_ store.Store, client *clientv3.Client) {
		ctx := context.Background()
		err := MigrateEnterpriseDB(ctx, client, migrations)
		if err != nil {
			t.Fatal(err)
		}
		version, err := GetEnterpriseDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 2; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
		if err := MigrateEnterpriseDB(ctx, client, migrations); err != nil {
			t.Fatal(err)
		}
		version, err = GetEnterpriseDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 2; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}
