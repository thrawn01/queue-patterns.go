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
Current Operating System has '10' CPUs
2024/09/17 17:42:09 INFO HTTP Listening ... address=127.0.0.1:2319
BenchmarkQueuePatterns/none
BenchmarkQueuePatterns/none-10         	      97	  11327316 ns/op	        88.28 ops/s
BenchmarkQueuePatterns/mutex
BenchmarkQueuePatterns/mutex-10        	     434	   2824829 ns/op	       354.0 ops/s
BenchmarkQueuePatterns/channel
BenchmarkQueuePatterns/channel-10      	     388	   2782659 ns/op	       359.4 ops/s
BenchmarkQueuePatterns/querator
BenchmarkQueuePatterns/querator-10     	     523	   2315109 ns/op	       431.9 ops/s
BenchmarkQueuePatterns/querator-noalloc
BenchmarkQueuePatterns/querator-noalloc-10         	     519	   2350627 ns/op	       425.4 ops/s
PASS
```

### TODO
- [ ] Channel (typical go style)
- [ ] Queue style Channels
- [ ] Ring Buffer
