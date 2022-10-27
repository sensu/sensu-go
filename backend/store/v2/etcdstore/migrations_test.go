package etcdstore_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
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

		// Retrieve all the newly created EntityConfig following the V2 -> V3
		// migration
		var entityConfigs []*corev3.EntityConfig
		req := storev2.NewResourceRequestFromResource(&corev3.EntityConfig{})
		list, err := s.List(ctx, req, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := list.UnwrapInto(&entityConfigs); err != nil {
			t.Fatal(err)
		}

		// For each EntityConfig, get its associated EntityState and combine
		// them into a v2.Entity
		var got []*corev2.Entity
		for _, entityConfig := range entityConfigs {
			entityState := &corev3.EntityState{Metadata: entityConfig.Metadata}
			stateReq := storev2.NewResourceRequestFromResource(entityState)

			wrappedState, err := s.Get(context.Background(), stateReq)
			if err != nil {
				t.Fatal(err)
			}
			wrappedState.UnwrapInto(entityState)

			v2Entity, err := corev3.V3EntityToV2(entityConfig, entityState)
			if err != nil {
				t.Fatal(err)
			}
			got = append(got, v2Entity)
		}

		// Those recreated v2.Entity should be exactly the same as the ones we
		// originally had in the database
		want := entities
		if len(got) != len(want) {
			t.Errorf("got %d entities, want %d", len(got), len(want))
		}
		for i, entity := range got {
			if !equalEntity(entity, want[i]) {
				t.Errorf("bad entity: got %v, want %v", entity, want[i])
			}
		}
	})
}

func equalEntity(got, want *corev2.Entity) bool {
	return proto.Equal(got, want)
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
