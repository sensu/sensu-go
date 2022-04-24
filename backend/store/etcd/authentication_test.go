package etcd

import (
	"bytes"
	"testing"
	"time"

	"github.com/sensu/sensu-go/backend/store"
	utilbytes "github.com/sensu/sensu-go/util/bytes"
)

func TestAuthenticationStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		// Secret is created
		secret, err := store.GetJWTSecret()
		if err != nil {
			t.Fatal(err)
		}

		if got, want := len(secret), 32; got != want {
			t.Fatalf("bad secret length: got %d, want %d", got, want)
		}

		time.Sleep(time.Second)

		// Retrieve the secret again, it should be the same
		result, err := store.GetJWTSecret()
		if err != nil {
			t.Fatal(err)
		}

		if got, want := result, secret; !bytes.Equal(got, want) {
			t.Errorf("bad secret result: got %x, want %x", got, want)
		}

		// We should be able to update it
		newSecret, err := utilbytes.Random(32)
		if err != nil {
			t.Fatal(err)
		}

		if err := store.UpdateJWTSecret(newSecret); err != nil {
			t.Fatal(err)
		}

		// The old and new secrets should not match
		result, err = store.GetJWTSecret()
		if err != nil {
			t.Fatal(err)
		}

		if got, want := result, secret; bytes.Equal(got, want) {
			t.Errorf("bad secret result: got %x, should not equal %x", got, want)
		}
		if got, want := result, newSecret; !bytes.Equal(got, want) {
			t.Errorf("bad secret result: got %x, want %x", got, want)
		}
	})
}
