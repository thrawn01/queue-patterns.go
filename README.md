This repo is intended to benchmark different queue patterns in golang.

### The Queue Pattern
Client needs to send lots of data to a Service. The naive design would be to
send each request as the need arises in an N+1 pattern. The more efficient
pattern would be to queue all the requests for some short period of time then
send all the items in a batch.

There are several different ways to implement a queue such that items are sent in a
batch. This repo is intended to benchmark those different approaches, so we
can evaluate each for performance and simplicity.


```
goos: darwin
goarch: arm64
pkg: github.com/thrawn01/queue-patterns.go
BenchmarkQueuePatterns
2024/09/17 15:51:50 INFO HTTP Listening ... address=127.0.0.1:2319
BenchmarkQueuePatterns/warmup
BenchmarkQueuePatterns/warmup-10                1023       1220565 ns/op           819.3 ops/s
BenchmarkQueuePatterns/mutex
BenchmarkQueuePatterns/mutex-10                  432       2933297 ns/op           340.9 ops/s
PASS
```

### TODO
- [ ] Channel (typical go style)
- [ ] Queue style Channels
- [ ] Ring Buffer
