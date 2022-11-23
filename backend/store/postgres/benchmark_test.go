package postgres

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func doQueryBatcherBenchmark(batchSize int, b *testing.B) {
	b.Helper()

	BatchTTL = 100 * time.Millisecond

	withPostgres(b, func(ctx context.Context, db *pgxpool.Pool, dsn string) {
		cfg := BatchConfig{
			DB:         db,
			BatchSize:  batchSize,
			Producers:  1000,
			Consumers:  1000,
			BufferSize: 10000,
		}
		batcher, err := NewEventBatcher(ctx, cfg)
		if err != nil {
			b.Fatal(err)
		}
		var i int64
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				n := atomic.AddInt64(&i, 1) % 1000
				args := EventArgs{
					Namespace:  "default",
					Check:      "check1",
					Entity:     fmt.Sprintf("entity-%d", n),
					LastOK:     12345,
					Status:     1,
					Serialized: []byte(`{"foo":"bar"}`),
					Selectors:  []byte("{}"),
				}
				_, err := batcher.Do(args)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

func BenchmarkQueryBatcher_1(b *testing.B) {
	doQueryBatcherBenchmark(1, b)
}

func BenchmarkQueryBatcher_2(b *testing.B) {
	doQueryBatcherBenchmark(2, b)
}
func BenchmarkQueryBatcher_3(b *testing.B) {
	doQueryBatcherBenchmark(3, b)
}
func BenchmarkQueryBatcher_4(b *testing.B) {
	doQueryBatcherBenchmark(4, b)
}
func BenchmarkQueryBatcher_5(b *testing.B) {
	doQueryBatcherBenchmark(5, b)
}
