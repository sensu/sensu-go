#! /bin/bash
docker stop pgbench
docker rm pgbench
docker build -t pgtest .
docker run -it -d -p 5432:5432 --name pgbench pgtest:latest
until (docker exec -it pgbench psql -c "SELECT 1"); do sleep 1; done;
docker exec -it pgbench psql -f /var/bench/schema.sql
docker exec -it pgbench psql -f /var/bench/seed.sql
docker exec -it pgbench pgbench -i
