This repo is intended to benchmark different queue patterns in golang.

### The Queue Pattern
A single writer needs to write lots of data. The naive design would be to
write each request as the need arises in an N+1 pattern. The more efficient
pattern would be to queue all the requests for some short period of time then
write all the items in a batch.

There are several different ways to implement a queue such that items are sent in a
batch. This repo is intended to benchmark those different approaches, so we
can evaluate each for performance and simplicity.


### MBP
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
BenchmarkQueuePatterns/querator-noalloc-10   519	   2350627 ns/op	       425.4 ops/s
PASS
```

### Server
```
Current Operating System has '128' CPUs
2024/09/17 18:58:15 INFO HTTP Listening ... address=127.0.0.1:2319
goos: linux
goarch: amd64
pkg: github.com/thrawn01/queue-patterns.go
cpu: AMD EPYC 7742 64-Core Processor
BenchmarkQueuePatterns/none-128         	     100	  10585102 ns/op	        94.46 ops/s	   15061 B/op	     124 allocs/op
BenchmarkQueuePatterns/mutex-128        	    5528	    212856 ns/op	      4698 ops/s	    3019 B/op	      11 allocs/op
BenchmarkQueuePatterns/channel-128      	    5554	    211581 ns/op	      4726 ops/s	    3034 B/op	      11 allocs/op
BenchmarkQueuePatterns/querator-128     	    7233	    167158 ns/op	      5982 ops/s	    2897 B/op	      12 allocs/op
BenchmarkQueuePatterns/querator-noalloc-128     7102	    168478 ns/op	      5935 ops/s	    2910 B/op	      12 allocs/op
PASS
```

### TODO
- [ ] Ring Buffer
