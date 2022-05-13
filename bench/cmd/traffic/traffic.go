package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/sensu/sensu-go/bench/cmd/traffic/silenced"
)

type arrayFlag []string

func (i *arrayFlag) String() string {
	return fmt.Sprint([]string(*i))
}

func (i *arrayFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var dsns arrayFlag
var connCount int
var strategy string

func main() {
	flag.Var(&dsns, "dsn", "Connection DSN")
	flag.IntVar(&connCount, "c", 0, "concurrent client connections")
	flag.StringVar(&strategy, "s", "counter", "traffic type")

	flag.Parse()

	if len(dsns) == 0 {
		log.Fatal("must configure at least one dsn. -dsn host=localhost")
	}
	if connCount < 1 {
		log.Fatal("must specify at least one connection. -c 1")
	}

	var worker Worker
	switch strategy {
	case "counter":
		worker = doCounterStuff
		log.Println("Using strategy: counter")
	case "silenced-rw-config":
		worker = silenced.ReadWriteConfig
		log.Println("Using strategy: silenced-rw-config")
	case "silenced-rw-discrete":
		worker = silenced.ReadWriteDiscrete
		log.Println("Using strategy: silenced-rw-discrete")
	default:
		log.Fatalf("Unknown strategy: %s. Must be one of counter, silenced-rw-config, silenced-rw-discrete", strategy)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-shutdown
		cancel()
	}()
	var connPool []*sql.DB
	for _, dsn := range dsns {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			log.Fatalf("connection open error: %v", err)
		}
		if err := db.PingContext(ctx); err != nil {
			log.Fatalf("Failed to establish connection '%s': %v", dsn, err)
		}
		connPool = append(connPool, db)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < connCount; i++ {
		wg.Add(1)
		go func(i int) {
			startWorker(ctx, connPool[i%len(connPool)], worker)
			wg.Done()
		}(i)

	}
	wg.Wait()
}

func startWorker(ctx context.Context, db *sql.DB, worker Worker) {
	randr := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		select {
		case <-ctx.Done():
			return
		default:
			txn, err := db.BeginTx(ctx, &sql.TxOptions{})
			if err != nil {
				log.Fatal(err)
			}
			if err := worker(ctx, randr, txn); err != nil {
				log.Printf("doit failed with error: %v\n", err)
				txn.Rollback()
			} else {
				txn.Commit()
			}
		}
	}

}

func doCounterStuff(ctx context.Context, randr *rand.Rand, tx *sql.Tx) error {
	op := randr.Intn(1000)
	tableSize := getTableSize(ctx, tx)
	rowid := randr.Int63n(tableSize)
	sleep := time.Duration(randr.Intn(5000))
	switch {
	case op > 975:
		_, err := tx.ExecContext(ctx, "UPDATE counters SET deleted_at = NOW() WHERE id = $1;", rowid)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO counters (c) VALUES (0);")
		if err != nil {
			return err
		}
	default:
		_, err := tx.ExecContext(ctx, "UPDATE counters SET c = c+1 WHERE id = $1 AND deleted_at IS NULL;", rowid)
		if err != nil {
			return err
		}
	}
	select {
	case <-ctx.Done():
	case <-time.After(time.Millisecond * sleep):
	}
	return nil
}

var (
	prevCt            int64
	ctRefreshDeadline time.Time
	mu                sync.Mutex
)

func getTableSize(ctx context.Context, tx *sql.Tx) int64 {
	if prevCt != 0 && time.Now().Before(ctRefreshDeadline) {
		return prevCt
	}
	mu.Lock()
	defer mu.Unlock()
	row := tx.QueryRowContext(ctx, "SELECT count(1) FROM counters")
	var count int64
	if err := row.Scan(&count); err != nil {
		log.Println("Failed to refresh counters table size")
		return prevCt
	}
	prevCt = count
	ctRefreshDeadline = time.Now().Add(time.Second * 30)
	return prevCt
}

type Worker func(context.Context, *rand.Rand, *sql.Tx) error
