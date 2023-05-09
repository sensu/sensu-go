package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/selector"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/stretchr/testify/assert"
)

const (
	defaultNamespace = "default"
	assetName        = "__test-asset__"
)

func testWithPostgresConfigStore(t testing.TB, fn func(p storev2.ConfigStore)) {
	t.Helper()
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		s := NewConfigStore(db)
		fn(s)
	})
}

func TestConfigStore_CreateOrUpdate(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		_, err := getAsset(ctx, s, "default", assetName)
		if err == nil {
			t.Error("expected non-nil error")
		}
		if _, ok := err.(*store.ErrNotFound); !ok {
			t.Errorf("wanted ErrNotFound, but got %T (%s)", err, err)
		}

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.ObjectMeta.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		var txInfo storev2.TxInfo

		ctx = storev2.ContextWithTxInfo(ctx, &txInfo)

		if err := createOrUpdateAsset(ctx, s, toCreate); err != nil {
			t.Error(err)
			return
		}

		if got, want := len(txInfo.Records), 1; got != want {
			t.Errorf("bad number of tx records: got %d, want %d", got, want)
			return
		}

		rec := txInfo.Records[0]

		if !rec.Created {
			t.Error("TxRecordInfo created flag not set")
		}

		if rec.ETag.Equals(nil) {
			t.Error("ETag not set")
		}
		if !rec.PrevETag.Equals(nil) {
			t.Error("PrevETag set")
		}

		etag := rec.ETag

		asset, err := getAsset(ctx, s, defaultNamespace, assetName)
		if err != nil {
			t.Error(err)
			return
		}

		assert.Equal(t, assetName, asset.Name)

		txInfo.Records = nil

		delete(toCreate.ObjectMeta.Labels, "label-0")
		delete(toCreate.ObjectMeta.Labels, "label-2")
		err = createOrUpdateAsset(ctx, s, toCreate)
		assert.NoError(t, err)

		if got, want := len(txInfo.Records), 1; got != want {
			t.Errorf("bad number of tx records: got %d, want %d", got, want)
			return
		}

		if !txInfo.Records[0].Updated {
			t.Error("TxRecordInfo updated flag not set")
		}

		if txInfo.Records[0].ETag.Equals(etag) {
			t.Error("different resource has the same etag")
		}

		asset, err = getAsset(ctx, s, "default", assetName)
		assert.NoError(t, err)
		assert.Equal(t, assetName, asset.Name)
	})
}

func TestConfigStore_CreateOrUpdateIfMatch(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		_, err := getAsset(ctx, s, "default", assetName)
		if err == nil {
			t.Error("expected non-nil error")
		}
		if _, ok := err.(*store.ErrNotFound); !ok {
			t.Errorf("wanted ErrNotFound, but got %T (%s)", err, err)
		}

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.ObjectMeta.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		var txInfo storev2.TxInfo

		ctx = storev2.ContextWithTxInfo(ctx, &txInfo)

		if err := createOrUpdateAsset(ctx, s, toCreate); err != nil {
			t.Error(err)
			return
		}

		if got, want := len(txInfo.Records), 1; got != want {
			t.Errorf("bad number of tx records: got %d, want %d", got, want)
			return
		}

		rec := txInfo.Records[0]

		if !rec.Created {
			t.Error("TxRecordInfo created flag not set")
		}

		if rec.ETag.Equals(nil) {
			t.Error("ETag not set")
		}
		if !rec.PrevETag.Equals(nil) {
			t.Error("PrevETag set")
		}

		etag := rec.ETag

		asset, err := getAsset(ctx, s, defaultNamespace, assetName)
		if err != nil {
			t.Error(err)
			return
		}

		assert.Equal(t, assetName, asset.Name)

		txInfo.Records = nil

		delete(toCreate.ObjectMeta.Labels, "label-0")
		delete(toCreate.ObjectMeta.Labels, "label-2")

		ctx = storev2.ContextWithIfMatch(ctx, []storev2.ETag{etag})
		err = createOrUpdateAsset(ctx, s, toCreate)
		if err == nil {
			t.Error("expected non-nil error")
			return
		}
		// IfMatch not supported
		if _, ok := err.(*store.ErrNotValid); !ok {
			t.Error(err)
			return
		}
	})
}

func TestConfigStore_CreateOrUpdateIfNoneMatch(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		_, err := getAsset(ctx, s, "default", assetName)
		if err == nil {
			t.Error("expected non-nil error")
		}
		if _, ok := err.(*store.ErrNotFound); !ok {
			t.Errorf("wanted ErrNotFound, but got %T (%s)", err, err)
		}

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.ObjectMeta.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		var txInfo storev2.TxInfo

		ctx = storev2.ContextWithTxInfo(ctx, &txInfo)

		if err := createOrUpdateAsset(ctx, s, toCreate); err != nil {
			t.Error(err)
			return
		}

		if got, want := len(txInfo.Records), 1; got != want {
			t.Errorf("bad number of tx records: got %d, want %d", got, want)
			return
		}

		rec := txInfo.Records[0]

		if !rec.Created {
			t.Error("TxRecordInfo created flag not set")
		}

		if rec.ETag.Equals(nil) {
			t.Error("ETag not set")
		}
		if !rec.PrevETag.Equals(nil) {
			t.Error("PrevETag set")
		}

		etag := rec.ETag

		asset, err := getAsset(ctx, s, defaultNamespace, assetName)
		if err != nil {
			t.Error(err)
			return
		}

		assert.Equal(t, assetName, asset.Name)

		txInfo.Records = nil

		delete(toCreate.ObjectMeta.Labels, "label-0")
		delete(toCreate.ObjectMeta.Labels, "label-2")

		ctx = storev2.ContextWithIfNoneMatch(ctx, []storev2.ETag{etag})
		err = createOrUpdateAsset(ctx, s, toCreate)
		if err == nil {
			t.Error("expected non-nil error")
			return
		}
		if err != storev2.ErrPreconditionFailed {
			t.Error(err)
			return
		}
		ctx = storev2.ContextWithIfNoneMatch(context.Background(), []storev2.ETag{storev2.ETag("asldkfjasldf")})
		err = createOrUpdateAsset(ctx, s, toCreate)
		if err != nil {
			t.Error(err)
		}
	})
}

func TestConfigStore_CreateIfNotExists(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		asset, err := getAsset(ctx, s, "default", assetName)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrNotFound{}, err)
		assert.Nil(t, asset)

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		var txInfo storev2.TxInfo
		ctx = storev2.ContextWithTxInfo(ctx, &txInfo)

		err = createIfNotExists(ctx, s, toCreate)
		assert.NoError(t, err)

		if got, want := len(txInfo.Records), 1; got != want {
			t.Errorf("bad number of tx records: got %d, want %d", got, want)
			return
		}

		rec := txInfo.Records[0]

		if !rec.Created {
			t.Error("TxRecordInfo created flag not set")
		}

		if rec.ETag.Equals(nil) {
			t.Error("ETag not set")
		}
		if !rec.PrevETag.Equals(nil) {
			t.Error("PrevETag set")
		}

		asset, err = getAsset(ctx, s, defaultNamespace, assetName)
		assert.NoError(t, err)
		assert.Equal(t, assetName, asset.Name)

		delete(toCreate.Labels, "label-0")
		delete(toCreate.Labels, "label-2")
		err = createIfNotExists(ctx, s, toCreate)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrAlreadyExists{}, err)

		// CreateIfNotExists will replace deleted resources
		if err := deleteAsset(ctx, s, "default", assetName); err != nil {
			t.Fatal(err)
		}
		err = createIfNotExists(ctx, s, toCreate)
		assert.NoError(t, err)

	})
}

func TestConfigStore_UpdateIfExists(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		asset, err := getAsset(ctx, s, "default", assetName)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrNotFound{}, err)
		assert.Nil(t, asset)

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err = updateIfExists(ctx, s, toCreate)
		if err == nil {
			t.Error("expected non-nil error")
		}
		if _, ok := err.(*store.ErrNotFound); !ok {
			t.Errorf("bad error type: got %T, want ErrNotFound (%s)", err, err)
		}

		err = createOrUpdateAsset(ctx, s, toCreate)
		assert.NoError(t, err)

		delete(toCreate.Labels, "label-0")
		delete(toCreate.Labels, "label-2")

		var txInfo storev2.TxInfo
		ctx = storev2.ContextWithTxInfo(ctx, &txInfo)

		if err := updateIfExists(ctx, s, toCreate); err != nil {
			t.Error(err)
			return
		}

		if got, want := len(txInfo.Records), 1; got != want {
			t.Errorf("bad number of tx records: got %d, want %d", got, want)
			return
		}

		rec := txInfo.Records[0]

		if rec.Created {
			t.Error("TxRecordInfo created flag set")
		}

		if !rec.Updated {
			t.Error("TxRecordInfo updated flag not set")
		}

		if rec.ETag.Equals(nil) {
			t.Error("ETag not set")
		}
		if rec.PrevETag.Equals(nil) {
			t.Error("PrevETag not set")
		}
		if rec.PrevETag.Equals(rec.ETag) {
			t.Error("PrevETag set incorrectly")
		}

		asset, err = getAsset(ctx, s, defaultNamespace, assetName)
		assert.NoError(t, err)
		assert.Equal(t, assetName, asset.Name)
	})
}

func TestConfigStore_UpdateIfExistsIfMatch(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		asset, err := getAsset(ctx, s, "default", assetName)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrNotFound{}, err)
		assert.Nil(t, asset)

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err = updateIfExists(ctx, s, toCreate)
		if err == nil {
			t.Error("expected non-nil error")
		}
		if _, ok := err.(*store.ErrNotFound); !ok {
			t.Errorf("bad error type: got %T, want ErrNotFound (%s)", err, err)
		}

		var txInfo storev2.TxInfo
		ctx = storev2.ContextWithTxInfo(ctx, &txInfo)

		err = createOrUpdateAsset(ctx, s, toCreate)
		assert.NoError(t, err)

		delete(toCreate.Labels, "label-0")
		delete(toCreate.Labels, "label-2")

		ifMatch := storev2.IfMatch{txInfo.Records[0].ETag}
		ctx = storev2.ContextWithIfMatch(ctx, ifMatch)

		if err := updateIfExists(ctx, s, toCreate); err != nil {
			t.Error(err)
			return
		}

		if got, want := len(txInfo.Records), 2; got != want {
			t.Errorf("bad number of tx records: got %d, want %d", got, want)
			return
		}

		rec := txInfo.Records[1]

		if rec.Created {
			t.Error("TxRecordInfo created flag set")
		}

		if !rec.Updated {
			t.Error("TxRecordInfo updated flag not set")
		}

		if rec.ETag.Equals(nil) {
			t.Error("ETag not set")
		}
		if rec.PrevETag.Equals(nil) {
			t.Error("PrevETag not set")
		}
		if rec.PrevETag.Equals(rec.ETag) {
			t.Error("PrevETag set incorrectly")
		}

		asset, err = getAsset(context.Background(), s, defaultNamespace, assetName)
		assert.NoError(t, err)
		assert.Equal(t, assetName, asset.Name)
	})
}

func TestConfigStore_UpdateIfExistsIfNoneMatch(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		asset, err := getAsset(ctx, s, "default", assetName)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrNotFound{}, err)
		assert.Nil(t, asset)

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err = updateIfExists(ctx, s, toCreate)
		if err == nil {
			t.Error("expected non-nil error")
		}
		if _, ok := err.(*store.ErrNotFound); !ok {
			t.Errorf("bad error type: got %T, want ErrNotFound (%s)", err, err)
		}

		var txInfo storev2.TxInfo
		ctx = storev2.ContextWithTxInfo(ctx, &txInfo)

		err = createOrUpdateAsset(ctx, s, toCreate)
		assert.NoError(t, err)

		delete(toCreate.Labels, "label-0")
		delete(toCreate.Labels, "label-2")

		ifNoneMatch := storev2.IfNoneMatch{txInfo.Records[0].ETag}
		ctx = storev2.ContextWithIfNoneMatch(ctx, ifNoneMatch)

		err = updateIfExists(ctx, s, toCreate)
		if err == nil {
			t.Error("expected non-nil error")
			return
		} else if err != storev2.ErrPreconditionFailed {
			t.Error(err)
			return
		}

		asset, err = getAsset(context.Background(), s, defaultNamespace, assetName)
		assert.NoError(t, err)
		assert.Equal(t, assetName, asset.Name)
	})
}

func TestConfigStore_Delete(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err := deleteAsset(ctx, s, defaultNamespace, assetName)
		if err == nil {
			t.Error("expected non-nil error")
		}
		if _, ok := err.(*store.ErrNotFound); !ok {
			t.Errorf("got %T, want ErrNotFound (%s)", err, err)
		}

		err = createOrUpdateAsset(ctx, s, toCreate)
		assert.NoError(t, err)

		_, err = getAsset(ctx, s, defaultNamespace, assetName)
		assert.NoError(t, err)

		err = deleteAsset(ctx, s, defaultNamespace, assetName)
		assert.NoError(t, err)
	})
}

func TestConfigStore_DeleteIfMatch(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err := createOrUpdateAsset(ctx, s, toCreate)
		if err != nil {
			t.Error(err)
			return
		}

		asset, err := getAsset(ctx, s, defaultNamespace, assetName)
		if err != nil {
			t.Error(err)
			return
		}

		etag, err := storev2.DecodeETag(asset.Annotations[store.SensuETagKey])
		if err != nil {
			t.Error(err)
			return
		}

		ctx = storev2.ContextWithIfMatch(ctx, storev2.IfMatch{storev2.ETag("asdflksd")})
		err = deleteAsset(ctx, s, defaultNamespace, assetName)
		if err == nil {
			t.Error("expected non-nil error")
			return
		}

		if err != storev2.ErrPreconditionFailed {
			t.Error(err)
			return
		}

		ctx = storev2.ContextWithIfMatch(context.Background(), storev2.IfMatch{etag})
		if err := deleteAsset(ctx, s, defaultNamespace, assetName); err != nil {
			t.Error(err)
		}
	})
}

func TestConfigStore_DeleteIfNoneMatch(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err := createOrUpdateAsset(ctx, s, toCreate)
		if err != nil {
			t.Error(err)
			return
		}

		asset, err := getAsset(ctx, s, defaultNamespace, assetName)
		if err != nil {
			t.Error(err)
			return
		}

		etag, err := storev2.DecodeETag(asset.Annotations[store.SensuETagKey])
		if err != nil {
			t.Error(err)
			return
		}

		ctx = storev2.ContextWithIfNoneMatch(context.Background(), storev2.IfNoneMatch{etag})
		err = deleteAsset(ctx, s, defaultNamespace, assetName)
		if err == nil {
			t.Error("expected non-nil error")
			return
		}

		if err != storev2.ErrPreconditionFailed {
			t.Error(err)
			return
		}

		ctx = storev2.ContextWithIfNoneMatch(context.Background(), storev2.IfNoneMatch{storev2.ETag("asdflksd")})

		err = deleteAsset(ctx, s, defaultNamespace, assetName)
		if err != nil {
			t.Error(err)
		}

	})
}

func TestConfigStore_Exists(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		exists, err := assetExists(ctx, s, defaultNamespace, assetName)
		assert.NoError(t, err)
		assert.False(t, exists)

		err = createOrUpdateAsset(ctx, s, toCreate)
		assert.NoError(t, err)

		exists, err = assetExists(ctx, s, defaultNamespace, assetName)
		assert.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestConfigStore_Get(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		asset, err := getAsset(ctx, s, "default", assetName)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrNotFound{}, err)
		assert.Nil(t, asset)

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err = createIfNotExists(ctx, s, toCreate)
		assert.NoError(t, err)

		asset, err = getAsset(ctx, s, defaultNamespace, assetName)
		assert.NoError(t, err)
		assert.Equal(t, assetName, asset.Name)
	})
}

func TestConfigStore_GetIfMatch(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		asset, err := getAsset(ctx, s, "default", assetName)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrNotFound{}, err)
		assert.Nil(t, asset)

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err = createIfNotExists(ctx, s, toCreate)
		assert.NoError(t, err)

		asset, err = getAsset(ctx, s, defaultNamespace, assetName)
		assert.NoError(t, err)
		assert.Equal(t, assetName, asset.Name)

		etag, err := storev2.DecodeETag(asset.Annotations[store.SensuETagKey])
		if err != nil {
			t.Error(err)
			return
		}
		ctx = storev2.ContextWithIfMatch(ctx, storev2.IfMatch{etag})

		asset, err = getAsset(ctx, s, defaultNamespace, assetName)
		if err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, assetName, asset.Name)

		ctx = storev2.ContextWithIfMatch(context.Background(), storev2.IfMatch{storev2.ETag("sldkfjsdlkf")})

		if _, err := getAsset(ctx, s, defaultNamespace, assetName); err == nil {
			t.Error("expected non-nil error")
		} else if err != storev2.ErrPreconditionFailed {
			t.Error("expected precondition to have failed")
		}
	})
}

func TestConfigStore_GetIfNoneMatch(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		asset, err := getAsset(ctx, s, "default", assetName)
		assert.Error(t, err)
		assert.IsType(t, &store.ErrNotFound{}, err)
		assert.Nil(t, asset)

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		err = createIfNotExists(ctx, s, toCreate)
		assert.NoError(t, err)

		asset, err = getAsset(ctx, s, defaultNamespace, assetName)
		assert.NoError(t, err)
		assert.Equal(t, assetName, asset.Name)

		etag, err := storev2.DecodeETag(asset.Annotations[store.SensuETagKey])
		if err != nil {
			t.Error(err)
			return
		}

		ctx = storev2.ContextWithIfNoneMatch(ctx, storev2.IfNoneMatch{etag})

		if _, err := getAsset(ctx, s, defaultNamespace, assetName); err == nil {
			t.Error("expected non-nil error")
		} else if err != storev2.ErrPreconditionFailed {
			t.Error("expected precondition to have failed")
		}

		ctx = storev2.ContextWithIfNoneMatch(context.Background(), storev2.IfNoneMatch{storev2.ETag("sdlfkjsdlfkj")})
		asset, err = getAsset(ctx, s, defaultNamespace, assetName)
		assert.NoError(t, err)
		assert.Equal(t, assetName, asset.Name)
	})
}

func TestConfigStore_List(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		assets, err := listAssets(ctx, s, defaultNamespace, &store.SelectionPredicate{})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(assets))

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		for i := 0; i < 100; i++ {
			toCreate.Name = fmt.Sprintf("%s%d__", assetName, i)
			err := createIfNotExists(ctx, s, toCreate)
			assert.NoError(t, err)
		}
		toCreate.Name = "__to_delete__"
		err = createIfNotExists(ctx, s, toCreate)
		assert.NoError(t, err)
		err = deleteAsset(ctx, s, "default", "__to_delete__")
		assert.NoError(t, err)

		for i := 0; i < 10; i++ {
			toCreate.Name = assetName
			toCreate.Namespace = fmt.Sprintf("ns%d", i)
			err := createIfNotExists(ctx, s, toCreate)
			assert.NoError(t, err)
		}

		assets, err = listAssets(ctx, s, defaultNamespace, &store.SelectionPredicate{})
		assert.NoError(t, err)
		if err != nil {
			return
		}
		assert.Equal(t, 100, len(assets))
		if len(assets) == 0 {
			return
		}

		t.Run("With Pagination", func(t *testing.T) {
			predicate := &store.SelectionPredicate{Limit: 45}

			assets, err = listAssets(ctx, s, defaultNamespace, predicate)
			assert.NoError(t, err)
			assert.Equal(t, 45, len(assets))
			assets, err = listAssets(ctx, s, defaultNamespace, predicate)
			assert.NoError(t, err)
			assert.Equal(t, 45, len(assets))
			assets, err = listAssets(ctx, s, defaultNamespace, predicate)
			assert.NoError(t, err)
			assert.Equal(t, 10, len(assets))
			assert.Equal(t, "", predicate.Continue)
		})

		assets, err = listAssets(ctx, s, "", &store.SelectionPredicate{})
		assert.NoError(t, err)
		assert.Equal(t, 110, len(assets))
	})
}

func TestConfigStore_List_WithSelectors(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		for i := 0; i < 100; i++ {
			toCreate := corev2.FixtureAsset(fmt.Sprintf("%s%d", assetName, i))
			toCreate.Labels[fmt.Sprintf("label-mod-key-%d", i%3)] = "value"
			toCreate.Labels["label-mod-value"] = fmt.Sprintf("value-%d", i%3)
			toCreate.Labels["label-flat"] = fmt.Sprintf("value-%d", i)
			toCreate.Labels["label-const"] = "const-value"
			toCreate.Sha512 = fmt.Sprintf("not-a-sha-%d", (i+2)%3)

			err := createIfNotExists(context.Background(), s, toCreate)
			assert.NoError(t, err)
		}
		tm := corev2.TypeMeta{
			APIVersion: "core/v2",
			Type:       "Asset",
		}

		tests := []struct {
			name               string
			selektor           *selector.Selector
			expectError        bool
			expectedAssetCount int
			expectedAssetNames []string
		}{
			{
				name: "asset name field and label -in- selector",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"label-flat", selector.InOperator, []string{"value-6", "value-22"}, selector.OperationTypeLabelSelector},
						{"asset.name", selector.InOperator, []string{assetName + "22", assetName + "45"}, selector.OperationTypeFieldSelector},
					},
				},
				expectError:        false,
				expectedAssetCount: 1,
				expectedAssetNames: []string{assetName + "22"},
			},
			{
				name: "label -in- selector",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"label-flat", selector.InOperator, []string{"value-6", "value-22"}, selector.OperationTypeLabelSelector},
					},
				},
				expectError:        false,
				expectedAssetCount: 2,
				expectedAssetNames: []string{assetName + "6", assetName + "22"},
			},
			{
				name: "asset name -in- selector",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"asset.name", selector.InOperator, []string{assetName + "6", assetName + "22"}, selector.OperationTypeFieldSelector},
					},
				},
				expectError:        false,
				expectedAssetCount: 2,
				expectedAssetNames: []string{assetName + "6", assetName + "22"},
			},
			{
				name: "asset name field -match- selector",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"asset.name", selector.MatchesOperator, []string{fmt.Sprintf("%s%d", assetName, 65)}, selector.OperationTypeFieldSelector},
					},
				},
				expectError:        false,
				expectedAssetCount: 1,
				expectedAssetNames: []string{assetName + "65"},
			},
			{
				name: "label -match- selector",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"label-flat", selector.MatchesOperator, []string{"value-65"}, selector.OperationTypeLabelSelector},
					},
				},
				expectError:        false,
				expectedAssetCount: 1,
				expectedAssetNames: []string{assetName + "65"},
			},
			{
				name: "field and label -match- selectors",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"asset.name", selector.MatchesOperator, []string{assetName + "6"}, selector.OperationTypeFieldSelector},
						{"label-mod-key-0", selector.MatchesOperator, []string{"value"}, selector.OperationTypeLabelSelector},
					},
				},
				expectError:        false,
				expectedAssetCount: 5,
				expectedAssetNames: []string{assetName + "6", assetName + "60", assetName + "63", assetName + "66", assetName + "69"},
			},
			{
				name: "field and label double equal selectors",
				selektor: &selector.Selector{
					Operations: []selector.Operation{
						{"asset.name", selector.DoubleEqualSignOperator, []string{assetName + "1"}, selector.OperationTypeFieldSelector},
						{"label-flat", selector.DoubleEqualSignOperator, []string{"value-1"}, selector.OperationTypeLabelSelector},
					},
				},
				expectError:        false,
				expectedAssetCount: 1,
				expectedAssetNames: []string{assetName + "1"},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				ctx := context.Background()
				selCtx := storev2.ContextWithSelector(ctx, tm, test.selektor)
				assets, err := listAssets(selCtx, s, "", &store.SelectionPredicate{})
				if test.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assetCount, err := countAssets(selCtx, s, defaultNamespace)
				if test.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, test.expectedAssetCount, len(assets))
				assert.Equal(t, test.expectedAssetCount, assetCount)
				for _, name := range test.expectedAssetNames {
					var found bool
					for _, asset := range assets {
						if asset.Name == name {
							found = true
							break
						}
					}
					assert.True(t, found, fmt.Sprintf("asset not found: %s", name))
				}
			})
		}
	})
}

func TestConfigStore_Count(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		ctx := context.Background()

		assets, err := listAssets(ctx, s, defaultNamespace, &store.SelectionPredicate{})
		assert.NoError(t, err)
		assert.Equal(t, 0, len(assets))

		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}

		for i := 0; i < 100; i++ {
			toCreate.Name = fmt.Sprintf("%s%d__", assetName, i)
			err := createIfNotExists(ctx, s, toCreate)
			assert.NoError(t, err)
		}

		for i := 0; i < 10; i++ {
			toCreate.Name = assetName
			toCreate.Namespace = fmt.Sprintf("ns%d", i)
			err := createIfNotExists(ctx, s, toCreate)
			assert.NoError(t, err)
		}

		ct, err := countAssets(ctx, s, defaultNamespace)
		assert.NoError(t, err)
		assert.Equal(t, 100, ct)

		ct, err = countAssets(ctx, s, "")
		assert.NoError(t, err)
		assert.Equal(t, 110, ct)
	})
}

func TestConfigStore_Patch(t *testing.T) {
	t.Skip("incomplete")
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		toCreate := corev2.FixtureAsset(assetName)
		for i := 0; i < 4; i++ {
			toCreate.Labels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("labelValue-%d", i)
		}
	})
}

func TestConfigStore_Watch(t *testing.T) {
	testWithPostgresConfigStore(t, func(s storev2.ConfigStore) {
		stor, ok := s.(*ConfigStore)
		if !ok {
			t.Error("expected config store")
			return
		}

		stor.watchInterval = time.Millisecond * 10
		stor.watchTxnWindow = time.Second

		ctx := context.Background()
		asset := corev2.FixtureAsset("my-asset")
		watchReq := storev2.ResourceRequest{
			APIVersion: "core/v2",
			Type:       "Asset",
		}
		watchChannel := s.Watch(ctx, watchReq)
		select {
		case record, ok := <-watchChannel:
			t.Errorf("expected watch channel to be empty. Got %v, %v", record, ok)
		default:
			// OK
		}

		// create notification
		err := createOrUpdateAsset(ctx, s, asset)
		if err != nil {
			t.Error(err)
			return
		}
		select {
		case watchEvents, ok := <-watchChannel:
			if !ok {
				t.Error("watcher closed unexpectedly")
				return
			}
			if len(watchEvents) != 1 {
				t.Error("expected 1 watch event")
				return
			}
			assert.Equal(t, storev2.WatchCreate, watchEvents[0].Type)

		case <-time.After(5 * time.Second):
			t.Fatalf("no watch event received before timeout")
		}

		// update notification
		asset.Labels["new-label"] = "new-value"
		err = createOrUpdateAsset(ctx, s, asset)
		if err != nil {
			t.Error(err)
			return
		}
		select {
		case watchEvents, ok := <-watchChannel:
			if !ok {
				t.Error("watcher closed unexpectedly")
				return
			}
			if len(watchEvents) != 1 {
				t.Error("expected 1 watch event")
				return
			}
			assert.Equal(t, storev2.WatchUpdate, watchEvents[0].Type)

		case <-time.After(5 * time.Second):
			t.Fatalf("no watch event received before timeout")
		}

		// delete notification
		err = deleteAsset(ctx, s, asset.Namespace, asset.Name)
		if err != nil {
			t.Error(err)
			return
		}
		if err != nil {
			t.Error(err)
			return
		}
		select {
		case watchEvents, ok := <-watchChannel:
			if !ok {
				t.Error("watcher closed unexpectedly")
				return
			}
			if len(watchEvents) != 1 {
				t.Error("expected 1 watch event")
				return
			}
			assert.Equal(t, storev2.WatchDelete, watchEvents[0].Type)

		case <-time.After(5 * time.Second):
			t.Fatalf("no watch event received before timeout")
		}
	})
}

func createOrUpdateAsset(ctx context.Context, pgStore storev2.ConfigStore, asset *corev2.Asset) error {
	req := storev2.ResourceRequest{
		APIVersion: "core/v2",
		Type:       "Asset",
		StoreName:  asset.StoreName(),
		Namespace:  asset.Namespace,
		Name:       asset.Name,
		SortOrder:  0,
	}

	wrapper, err := wrapAsset(asset)
	if err != nil {
		return err
	}

	return pgStore.CreateOrUpdate(ctx, req, wrapper)
}

func createIfNotExists(ctx context.Context, pgStore storev2.ConfigStore, asset *corev2.Asset) error {
	req := storev2.ResourceRequest{
		APIVersion: "core/v2",
		Type:       "Asset",
		StoreName:  asset.StoreName(),
		Namespace:  asset.Namespace,
		Name:       asset.Name,
		SortOrder:  0,
	}

	wrapper, err := wrapAsset(asset)
	if err != nil {
		return err
	}

	return pgStore.CreateIfNotExists(ctx, req, wrapper)
}

func countAssets(ctx context.Context, pgStore storev2.ConfigStore, namespace string) (int, error) {
	asset := &corev2.Asset{}
	req := storev2.ResourceRequest{
		Namespace:  namespace,
		StoreName:  asset.StoreName(),
		APIVersion: "core/v2",
		Type:       "Asset",
	}

	return pgStore.Count(ctx, req)
}

func listAssets(ctx context.Context, pgStore storev2.ConfigStore, namespace string, predicate *store.SelectionPredicate) ([]*corev2.Asset, error) {
	asset := &corev2.Asset{}
	req := storev2.ResourceRequest{
		Namespace:  namespace,
		Name:       "",
		StoreName:  asset.StoreName(),
		APIVersion: "core/v2",
		Type:       "Asset",
		SortOrder:  0,
	}

	list, err := pgStore.List(ctx, req, predicate)
	if err != nil {
		return nil, err
	}

	res, err := list.Unwrap()
	if err != nil {
		return nil, err
	}

	assets := make([]*corev2.Asset, 0, len(res))
	for _, a := range res {
		asset, ok := a.(*corev2.Asset)
		if !ok {
			return nil, errors.New("not an asset")
		}
		assets = append(assets, asset)
	}

	return assets, nil
}

func getAsset(ctx context.Context, pgStore storev2.ConfigStore, namespace, name string) (*corev2.Asset, error) {
	asset := &corev2.Asset{}
	req := storev2.ResourceRequest{
		Namespace:  namespace,
		Name:       name,
		StoreName:  asset.StoreName(),
		APIVersion: "core/v2",
		Type:       "Asset",
		SortOrder:  0,
	}

	assetWrapper, err := pgStore.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	res, err := assetWrapper.Unwrap()
	if err != nil {
		return nil, err
	}

	asset, ok := res.(*corev2.Asset)
	if !ok {
		return nil, errors.New("resource is not an asset")
	}

	return asset, nil
}

func deleteAsset(ctx context.Context, pgStore storev2.ConfigStore, namespace, name string) error {
	asset := &corev2.Asset{}
	req := storev2.ResourceRequest{
		Namespace:  namespace,
		Name:       name,
		StoreName:  asset.StoreName(),
		APIVersion: "core/v2",
		Type:       "Asset",
		SortOrder:  0,
	}

	return pgStore.Delete(ctx, req)
}

func assetExists(ctx context.Context, pgStore storev2.ConfigStore, namespace, name string) (bool, error) {
	asset := &corev2.Asset{}
	req := storev2.ResourceRequest{
		Namespace:  namespace,
		Name:       name,
		StoreName:  asset.StoreName(),
		APIVersion: "core/v2",
		Type:       "Asset",
		SortOrder:  0,
	}

	return pgStore.Exists(ctx, req)
}

func updateIfExists(ctx context.Context, pgStore storev2.ConfigStore, asset *corev2.Asset) error {
	req := storev2.ResourceRequest{
		APIVersion: "core/v2",
		Type:       "Asset",
		StoreName:  asset.StoreName(),
		Namespace:  asset.Namespace,
		Name:       asset.Name,
		SortOrder:  0,
	}

	wrapper, err := wrapAsset(asset)
	if err != nil {
		return err
	}

	return pgStore.UpdateIfExists(ctx, req, wrapper)
}

func wrapAsset(asset *corev2.Asset) (*wrap.Wrapper, error) {
	jsonAsset, err := json.Marshal(asset)
	if err != nil {
		return nil, err
	}

	return &wrap.Wrapper{
		TypeMeta: &corev2.TypeMeta{
			APIVersion: "core/v2",
			Type:       "Asset",
		},
		Encoding:    wrap.Encoding_json,
		Compression: wrap.Compression_none,
		Value:       jsonAsset,
	}, nil
}

func TestConfigStore_UpdatedSince(t *testing.T) {
	withPostgres(t, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		s := &ConfigStore{
			db: db,
		}
		older := corev2.FixtureAsset("older")
		err := createIfNotExists(ctx, s, older)
		if err != nil {
			t.Error(err)
			return
		}
		wrapper, err := s.Get(ctx, storev2.NewResourceRequestFromResource(older))
		if err != nil {
			t.Error(err)
			return
		}
		var asset corev2.Asset
		if err := wrapper.UnwrapInto(&asset); err != nil {
			t.Error(err)
			return
		}
		updatedSince := asset.ObjectMeta.Labels[store.SensuUpdatedAtKey]
		if updatedSince == "" {
			t.Error("no updated_since attribute")
			return
		}
		pred := &store.SelectionPredicate{
			UpdatedSince: updatedSince,
		}
		time.Sleep(2 * time.Second) // ensure that "newer" is at least one second older than "older"
		newer := corev2.FixtureAsset("newer")
		err = createIfNotExists(ctx, s, newer)
		if err != nil {
			t.Error(err)
			return
		}
		list, err := s.List(ctx, storev2.NewResourceRequestFromResource(&asset), pred)
		if err != nil {
			t.Error(err)
			return
		}
		var assets []*corev2.Asset
		if err := list.UnwrapInto(&assets); err != nil {
			t.Error(err)
			return
		}
		if len(assets) != 1 {
			t.Errorf("wrong number of assets: want 1, got %d", len(assets))
			return
		}
		if got, want := assets[0].ObjectMeta.Name, "newer"; got != want {
			t.Errorf("bad asset name: got %q, want %q", got, want)
		}
	})
}
