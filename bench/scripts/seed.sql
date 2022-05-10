-- insert 5k counters
INSERT INTO COUNTERS (c) SELECT 0 FROM generate_series(1, 5000);
