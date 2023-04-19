//go:build integration
// +build integration

package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/echlebek/crock"
	time "github.com/echlebek/timeproxy"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/etcd"
	store "github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/types"
)

var mockTime = crock.NewTime(time.Unix(0, 0))

func init() {
	mockTime.Resolution = time.Millisecond
	mockTime.Multiplier = 100
	time.TimeProxy = mockTime
}

func TestResourceCacheIntegration(t *testing.T) {
	mockTime.Start()
	defer mockTime.Stop()
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()

	client := e.NewEmbeddedClient()

	store := store.NewStore(client, e.Name())

	if err := store.CreateNamespace(context.Background(), corev2.FixtureNamespace("default")); err != nil {
		t.Fatal(err)
	}
	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, "default")
	fixtures := []Value{}

	// Populate store with some initial checks
	// v2 entities are no longer compatible with the v2 cache
	for i := 0; i < 9; i++ {
		fixture := corev2.FixtureCheckConfig(fmt.Sprintf("%d", i))
		fixture.Command = "test"
		fixtures = append(fixtures, getCacheValue(fixture, true))
		if err := store.UpdateCheckConfig(ctx, fixture); err != nil {
			t.Fatal(err)
		}
	}

	otherFixtures := []Value{}

	// Include some checks from a non-default namespace
	if err := store.CreateNamespace(context.Background(), &corev2.Namespace{Name: "other"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		fixture := corev2.FixtureCheckConfig(fmt.Sprintf("%d", i))
		fixture.Namespace = "other"
		fixture.Command = "test"
		otherFixtures = append(otherFixtures, getCacheValue(fixture, true))
		if err := store.UpdateCheckConfig(ctx, fixture); err != nil {
			t.Fatal(err)
		}
	}

	cacheCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache, err := New(cacheCtx, client, &corev2.CheckConfig{}, true)
	if err != nil {
		t.Fatal(err)
	}

	watcher := cache.Watch(cacheCtx)

	if got, want := cache.Get("default"), fixtures; !checkResources(t, got, want) {
		t.Fatalf("bad resources")
	}

	if got, want := cache.Get("notdefault"), []Value{}; !checkResources(t, got, want) {
		t.Fatal("bad resources")
	}

	if got, want := cache.Get("other"), otherFixtures; !checkResources(t, got, want) {
		t.Fatal("bad resources")
	}

	newCheck := corev2.FixtureCheckConfig("new")
	newCheck.Command = "test"
	if err := store.UpdateCheckConfig(ctx, newCheck); err != nil {
		t.Fatal(err)
	}
	<-watcher

	got := cache.Get("default")

	if got, want := got[len(got)-1], getCacheValue(newCheck, true); got.Resource.GetObjectMeta().Name != want.Resource.GetObjectMeta().Name {
		t.Errorf("bad resource: got %s, want %s", got.Resource.GetObjectMeta().Name, want.Resource.GetObjectMeta().Name)
	}

	if err := store.DeleteCheckConfigByName(ctx, newCheck.Name); err != nil {
		t.Fatal(err)
	}

	<-watcher

	if got, want := cache.Get("default"), fixtures; !checkResources(t, got, want) {
		t.Errorf("bad resources")
	}

}

func checkResources(t testing.TB, got, want []Value) bool {
	t.Helper()
	success := true
	if got, want := len(got), len(want); got != want {
		t.Errorf("lengths do not match: got %d, want %d", got, want)
		return false
	}
	for i := range got {
		if got, want := got[i].Resource.GetObjectMeta(), want[i].Resource.GetObjectMeta(); got.Cmp(&want) != 0 {
			t.Errorf("value %d: got %v, want %v", i, got, want)
			success = false
		}
	}
	return success
}
