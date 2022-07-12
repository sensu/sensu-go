package etcdstore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func testMigration(ctx context.Context, client *clientv3.Client) error {
	return nil
}

func testFailingMigration(ctx context.Context, client *clientv3.Client) error {
	return errors.New("error")
}

func TestMigrateDBBase(t *testing.T) {
	migrations := []etcdstore.Migration{etcdstore.Base}
	testWithEtcdClient(t, func(_ storev2.Interface, client *clientv3.Client) {
		ctx := context.Background()
		err := etcdstore.MigrateDB(ctx, client, migrations)
		if err != nil {
			t.Fatal(err)
		}
		version, err := etcdstore.GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 0; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestMigrationDBBroken(t *testing.T) {
	migrations := []etcdstore.Migration{
		etcdstore.Base,
		testFailingMigration,
		testMigration,
	}
	testWithEtcdClient(t, func(_ storev2.Interface, client *clientv3.Client) {
		ctx := context.Background()
		err := etcdstore.MigrateDB(ctx, client, migrations)
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		version, err := etcdstore.GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 0; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestMigrationDBLastBroken(t *testing.T) {
	migrations := []etcdstore.Migration{
		etcdstore.Base,
		testMigration,
		testFailingMigration,
	}
	testWithEtcdClient(t, func(_ storev2.Interface, client *clientv3.Client) {
		ctx := context.Background()
		err := etcdstore.MigrateDB(ctx, client, migrations)
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		version, err := etcdstore.GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 1; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestMigrationDBGood(t *testing.T) {
	migrations := []etcdstore.Migration{
		etcdstore.Base,
		testMigration,
		testMigration,
	}
	testWithEtcdClient(t, func(_ storev2.Interface, client *clientv3.Client) {
		ctx := context.Background()
		err := etcdstore.MigrateDB(ctx, client, migrations)
		if err != nil {
			t.Fatal(err)
		}
		version, err := etcdstore.GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 2; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestReapplyMigration(t *testing.T) {
	migrations := []etcdstore.Migration{
		etcdstore.Base,
		testMigration,
		testMigration,
	}
	testWithEtcdClient(t, func(_ storev2.Interface, client *clientv3.Client) {
		ctx := context.Background()
		err := etcdstore.MigrateDB(ctx, client, migrations)
		if err != nil {
			t.Fatal(err)
		}
		version, err := etcdstore.GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 2; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
		if err := etcdstore.MigrateDB(ctx, client, migrations); err != nil {
			t.Fatal(err)
		}
		version, err = etcdstore.GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 2; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestUpgradeV2Entities(t *testing.T) {
	migrations := []etcdstore.Migration{
		etcdstore.Base,
		etcdstore.MigrateV2EntityToV3,
	}
	testWithEtcdClient(t, func(s storev2.Interface, client *clientv3.Client) {
		ctx := store.NamespaceContext(context.Background(), "default")
		entities := make([]*corev2.Entity, 10)
		for i := range entities {
			entities[i] = corev2.FixtureEntity(fmt.Sprintf("%d", i))
		}
		for _, entity := range entities {
			req := storev2.NewResourceRequestFromResource(entity)
			wrapper, err := storev2.WrapResource(entity)
			if err != nil {
				t.Fatal(err)
			}
			if err := s.CreateIfNotExists(ctx, req, wrapper); err != nil {
				t.Fatal(err)
			}
		}
		if err := etcdstore.MigrateDB(ctx, client, migrations); err != nil {
			t.Fatal(err)
		}
		var got []*corev2.Entity
		req := storev2.NewResourceRequestFromResource(&corev2.Entity{})
		wrapped, err := s.List(ctx, req, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := wrapped.UnwrapInto(&entities); err != nil {
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
	migrations := []etcdstore.Migration{etcdstore.Base}
	testWithEtcdClient(t, func(_ storev2.Interface, client *clientv3.Client) {
		ctx := context.Background()
		err := etcdstore.MigrateEnterpriseDB(ctx, client, migrations)
		if err != nil {
			t.Fatal(err)
		}
		version, err := etcdstore.GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 0; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestMigrationDBBrokenEnterprise(t *testing.T) {
	migrations := []etcdstore.Migration{
		etcdstore.Base,
		testFailingMigration,
		testMigration,
	}
	testWithEtcdClient(t, func(_ storev2.Interface, client *clientv3.Client) {
		ctx := context.Background()
		err := etcdstore.MigrateEnterpriseDB(ctx, client, migrations)
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		version, err := etcdstore.GetDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 0; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestMigrationDBLastBrokenEnterprise(t *testing.T) {
	migrations := []etcdstore.Migration{
		etcdstore.Base,
		testMigration,
		testFailingMigration,
	}
	testWithEtcdClient(t, func(_ storev2.Interface, client *clientv3.Client) {
		ctx := context.Background()
		err := etcdstore.MigrateEnterpriseDB(ctx, client, migrations)
		if err == nil {
			t.Fatal("expected non-nil error")
		}
		version, err := etcdstore.GetEnterpriseDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 1; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestMigrationDBGoodEnterprise(t *testing.T) {
	migrations := []etcdstore.Migration{
		etcdstore.Base,
		testMigration,
		testMigration,
	}
	testWithEtcdClient(t, func(_ storev2.Interface, client *clientv3.Client) {
		ctx := context.Background()
		err := etcdstore.MigrateEnterpriseDB(ctx, client, migrations)
		if err != nil {
			t.Fatal(err)
		}
		version, err := etcdstore.GetEnterpriseDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 2; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}

func TestReapplyMigrationEnterprise(t *testing.T) {
	migrations := []etcdstore.Migration{
		etcdstore.Base,
		testMigration,
		testMigration,
	}
	testWithEtcdClient(t, func(_ storev2.Interface, client *clientv3.Client) {
		ctx := context.Background()
		err := etcdstore.MigrateEnterpriseDB(ctx, client, migrations)
		if err != nil {
			t.Fatal(err)
		}
		version, err := etcdstore.GetEnterpriseDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 2; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
		if err := etcdstore.MigrateEnterpriseDB(ctx, client, migrations); err != nil {
			t.Fatal(err)
		}
		version, err = etcdstore.GetEnterpriseDatabaseVersion(ctx, client)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := version, 2; got != want {
			t.Errorf("bad database version: got %d, want %d", got, want)
		}
	})
}
