package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v4/pgxpool"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
)

func testWithSilenceStore(t testing.TB, fn func(*SilenceStore, *NamespaceStore)) {
	t.Helper()

	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		silenceStore := NewSilenceStore(db)
		nsStore := NewNamespaceStore(db)

		namespace := &corev3.Namespace{
			Metadata: &corev2.ObjectMeta{
				Name:        "default",
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
			},
		}
		if err := nsStore.CreateIfNotExists(ctx, namespace); err != nil {
			panic(err)
			t.Fatal(err)
		}

		fn(silenceStore, nsStore)
	})
}

func TestSilenceStoreCreateOrUpdate(t *testing.T) {
	goodSilence := &corev2.Silenced{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "silence",
		},
		Check:           "foo",
		Subscription:    "bar",
		Reason:          "none",
		ExpireOnResolve: true,
		Begin:           time.Now().Unix(),
		ExpireAt:        time.Now().Add(time.Minute).Unix(),
	}
	badSilence := &corev2.Silenced{
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "asdf", // namespace DNE
			Name:      "silence",
		},
		Check:           "foo",
		Subscription:    "bar",
		Reason:          "none",
		ExpireOnResolve: true,
		Begin:           time.Now().Unix(),
		ExpireAt:        time.Now().Add(time.Minute).Unix(),
	}
	ctx := context.Background()
	testWithSilenceStore(t, func(sstore *SilenceStore, nsStore *NamespaceStore) {
		if err := sstore.UpdateSilence(ctx, goodSilence); err != nil {
			t.Fatal(err)
		}
		if err := sstore.UpdateSilence(ctx, badSilence); err == nil {
			t.Fatal("expected non-nil error")
		} else if nserr, ok := err.(*store.ErrNamespaceMissing); ok {
			if got, want := nserr.Namespace, "asdf"; got != want {
				t.Errorf("namespace in error incorrect: got %s, want %s", got, want)
			}
		} else {
			t.Fatal(err)
		}
	})
}

func TestSilenceStoreGetSilences(t *testing.T) {
	want := []*corev2.Silenced{
		corev2.FixtureSilenced("foo:bar"),
		corev2.FixtureSilenced("bar:baz"),
		corev2.FixtureSilenced("baz:foo"),
	}
	ctx := context.Background()
	testWithSilenceStore(t, func(sstore *SilenceStore, nsStore *NamespaceStore) {
		for _, silence := range want {
			if err := sstore.UpdateSilence(ctx, silence); err != nil {
				t.Fatal(err)
			}
		}
		got, err := sstore.GetSilences(ctx, "default")
		if err != nil {
			t.Fatal(err)
		}
		if !cmp.Equal(got, want) {
			t.Errorf("silences not equal: got %v", cmp.Diff(got, want))
		}
		got, err = sstore.GetSilences(ctx, "alsdkjf")
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(got), 0; got != want {
			t.Errorf("got more than %d silences: %d", want, got)
		}
	})
}

func TestSilenceStoreGetSilencesByCheck(t *testing.T) {
	silences := []*corev2.Silenced{
		corev2.FixtureSilenced("foo:bar"),
		corev2.FixtureSilenced("bar:bar"),
		corev2.FixtureSilenced("baz:foo"),
	}

	want := silences[:2]

	ctx := context.Background()
	testWithSilenceStore(t, func(sstore *SilenceStore, nsStore *NamespaceStore) {
		for _, silence := range silences {
			if err := sstore.UpdateSilence(ctx, silence); err != nil {
				t.Fatal(err)
			}
		}
		got, err := sstore.GetSilencesByCheck(ctx, "default", "bar")
		if err != nil {
			t.Fatal(err)
		}
		if !cmp.Equal(got, want) {
			t.Errorf("silences not equal: got %v", cmp.Diff(got, want))
		}
	})
}

func TestSilenceStoreGetSilencesBySubscription(t *testing.T) {
	silences := []*corev2.Silenced{
		corev2.FixtureSilenced("foo:bar"),
		corev2.FixtureSilenced("bar:bar"),
		corev2.FixtureSilenced("baz:foo"),
	}

	want := silences[:2]

	ctx := context.Background()
	testWithSilenceStore(t, func(sstore *SilenceStore, nsStore *NamespaceStore) {
		for _, silence := range silences {
			if err := sstore.UpdateSilence(ctx, silence); err != nil {
				t.Fatal(err)
			}
		}
		got, err := sstore.GetSilencesBySubscription(ctx, "default", []string{"foo", "bar"})
		if err != nil {
			t.Fatal(err)
		}
		if !cmp.Equal(got, want) {
			t.Errorf("silences not equal: got %v", cmp.Diff(got, want))
		}
	})
}

func TestSilenceStoreGetSilenceByName(t *testing.T) {
	silences := []*corev2.Silenced{
		corev2.FixtureSilenced("foo:bar"),
		corev2.FixtureSilenced("bar:bar"),
		corev2.FixtureSilenced("baz:foo"),
	}

	want := silences[0]

	ctx := context.Background()
	testWithSilenceStore(t, func(sstore *SilenceStore, nsStore *NamespaceStore) {
		for _, silence := range silences {
			if err := sstore.UpdateSilence(ctx, silence); err != nil {
				t.Fatal(err)
			}
		}
		got, err := sstore.GetSilenceByName(ctx, "default", "foo:bar")
		if err != nil {
			t.Fatal(err)
		}
		if !cmp.Equal(got, want) {
			t.Errorf("silences not equal: got %v", cmp.Diff(got, want))
		}
	})
}

func TestSilenceStoreGetSilencesByName(t *testing.T) {
	silences := []*corev2.Silenced{
		corev2.FixtureSilenced("foo:bar"),
		corev2.FixtureSilenced("bar:bar"),
		corev2.FixtureSilenced("baz:foo"),
	}

	want := silences[:2]

	ctx := context.Background()
	testWithSilenceStore(t, func(sstore *SilenceStore, nsStore *NamespaceStore) {
		for _, silence := range silences {
			if err := sstore.UpdateSilence(ctx, silence); err != nil {
				t.Fatal(err)
			}
		}
		got, err := sstore.GetSilencesByName(ctx, "default", []string{"foo:bar", "bar:bar"})
		if err != nil {
			t.Fatal(err)
		}
		if !cmp.Equal(got, want) {
			t.Errorf("silences not equal: got %v", cmp.Diff(got, want))
		}
	})
}
