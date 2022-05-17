package postgres

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

func doQueryBatcherBenchmark(batchSize int, b *testing.B) {
	BatchTTL = 100 * time.Millisecond
	pgURL := os.Getenv("PG_URL")
	if pgURL == "" {
		b.Skip("skipping postgres test")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, err := pgxpool.Connect(ctx, pgURL)
	if err != nil {
		b.Fatal(err)
	}
	dbName := "sensu" + strings.ReplaceAll(uuid.New().String(), "-", "")
	if _, err := db.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", dbName)); err != nil {
		b.Fatal(err)
	}
	defer dropAll(ctx, dbName, pgURL)
	db.Close()
	db, err = pgxpool.Connect(ctx, fmt.Sprintf("dbname=%s ", dbName)+pgURL)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	if err := upgrade(ctx, db); err != nil {
		b.Fatal(err)
	}
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
