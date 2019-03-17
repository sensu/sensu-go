[![Build Status](https://api.travis-ci.org/sensu/lasr.svg)](https://api.travis-ci.org/sensu/lasr)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/sensu/lasr)

# lasr
A persistent message queue backed by BoltDB. This queue is useful when the producers and consumers can live in the same process.

Project goals
-------------
  * Data integrity over performance.
  * Simplicity over complexity.
  * Ease of use.
  * Minimal feature set.

Safety
------
lasr is designed to never lose information. When the Send method completes, messages have been safely written to disk. On Receive, messages are not deleted until Ack is called. Users should make sure they always respond to messages with Ack or Nack.

Misc
----
Dead-lettering is supported, but disabled by default.

Benchmarks
----------

On 5th Gen Lenovo X1 Carbon with 512 GB SSD:

`$ hey -m POST -D main.go -h2 -cpus 2 -n 20000 -c 10 http://localhost:8080`

```
Summary:
  Total:        1.8671 secs
  Slowest:      0.0112 secs
  Fastest:      0.0001 secs
  Average:      0.0009 secs
  Requests/sec: 10711.7919

Response time histogram:
  0.000 [1]     |
  0.001 [14044] |∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.002 [5030]  |∎∎∎∎∎∎∎∎∎∎∎∎∎∎
  0.003 [709]   |∎∎
  0.005 [139]   |
  0.006 [36]    |
  0.007 [19]    |
  0.008 [8]     |
  0.009 [7]     |
  0.010 [3]     |
  0.011 [4]     |

Latency distribution:
  10% in 0.0001 secs
  25% in 0.0003 secs
  50% in 0.0008 secs
  75% in 0.0013 secs
  90% in 0.0018 secs
  95% in 0.0022 secs
  99% in 0.0034 secs

Details (average, fastest, slowest):
  DNS+dialup: 0.0000 secs, 0.0000 secs, 0.0056 secs
  DNS-lookup: 0.0000 secs, 0.0000 secs, 0.0020 secs
  req write:  0.0000 secs, 0.0000 secs, 0.0042 secs
  resp wait:  0.0009 secs, 0.0000 secs, 0.0098 secs
  resp read:  0.0000 secs, 0.0000 secs, 0.0038 secs

Status code distribution:
  [200]	20000 responses
```
