# Running

1. build dockerfile
    `docker build -t pgtest .`
1. start container
    `docker run -it -d -p 5432:5432 --name pgbench pgtest:latest`
1. Apply DB Schema
    `docker exec -it pgbench psql -f /var/bench/schema.sql`
1. Seed Data
    `docker exec -it pgbench psql -f /var/bench/seed.sql`
1. Start a watcher on the counters table. (runs until interrupted)
    Locally:
    `go run ./cmd/watch -dsn "host=localhost user=postgres password=password sslmode=disable"`
    Or in container:
    `docker exec -it pgbench watch -dsn "sslmode=disable"`
1. Start traffic tool to generate traffic. (runs until interrupted)
    `docker exec -it pgbench traffic -dsn "sslmode=disable" -c 100`
1. After sufficient traffic has been genrated, kill the traffic THEN the watcher process



## Silenced perf comparison

1. rebuild the environment
    `./rebuild.sh`
1. Add silenced load to the configuration table (runs until interrupted)
    `docker exec -it pgbench traffic -dsn sslmode=disable -c 100 -s silenced-rw-config`
1. Run pgbench to measuer perf of random config reads
    `pgbench -f scripts/random_config_access.sql -c 50 -t 10000 -U postgres -h localhost postgres`
1. Stop the silenced config traffic and start silenced discrete table traffic
    `docker exec -it pgbench traffic -dsn sslmode=disable -c 100 -s silenced-rw-discrete`
1. Rerun pgbench and compare results between scenarios.
