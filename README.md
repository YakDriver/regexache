# regexache

`regexache` is a thread-safe regular expression cache, providing a drop-in replacement for `regexp.MustCompile()` (`regexache` calls `regexp.MustCompile()` on your behalf to populate the cache). This special purpose cache specifically addresses regular expressions, which use a lot of memory. In a [project](https://github.com/hashicorp/terraform-provider-aws) with about ~4500 regexes, using `regexache` saved nearly 20% total memory use.

Unlike excellent caches, such as [go-cache](https://github.com/patrickmn/go-cache) or memcached, the calling code does not need to know anything about the cache or instantiate it, simply dropping in `regexache.MustCompile()` in place of `regexp.MustCompile()`. There are cons to this approach but for an existing large project, they may be outweighed by not needing to rework existing code (other than the drop in).

For projects with few regular expressions, caching is unlikely to improve memory use--stick with static use of `regexp.MustCompile()`. For projects with thousands of regular expressions, and especially untracked duplicates, using `regexache` can save significant memory.

Potential problems with using `regexache` include cache contention and preventing garbage collection of regular expressions. Cache contention results from the cache map being read-locked for reads and locked for updates. For garbage collection, if you're not using `regexache` and instantiate a regular expressions locally and it goes out of scope without any references to it remaining, Go may reclaim the memory. However, `regexache` keeps pointers to the regular expressions in the cache so they cannot be garbage collected until the entry expires and is cleaned out of the cache. Benchmark various expiration settings to see what works best.

## Using regexache

Using `regexache` is simple. If this is your code before, see below for code after.

Before `regexache`:

```go
package main

import (
	"fmt"
	"regexp"
)

func main() {
	var validID = regexp.MustCompile(`^[a-z]+\[[0-9]+\]$`)

	fmt.Println(validID.MatchString("adam[23]"))
	fmt.Println(validID.MatchString("eve[7]"))
	fmt.Println(validID.MatchString("Job[48]"))
	fmt.Println(validID.MatchString("snakey"))
}
```
([Playground](https://go.dev/play/p/e0MHgtJFNHE))

After `regexache`:

```go
package main

import (
	"fmt"

	"github.com/YakDriver/regexache"
)

func main() {
	var validID = regexache.MustCompile(`^[a-z]+\[[0-9]+\]$`)

	fmt.Println(validID.MatchString("adam[23]"))
	fmt.Println(validID.MatchString("eve[7]"))
	fmt.Println(validID.MatchString("Job[48]"))
	fmt.Println(validID.MatchString("snakey"))
}
```
([Playground](https://go.dev/play/p/q0apcbfeMV-))


## Environment Variables

| Env Var | Description |
| --- | --- |
| REGEXACHE_OFF | Any value will turn `regexache` completely off. All other environment variables are ignored. Useful for testing with and without caching. When off, `regexache.MustCompile()` is equivalent to `regexp.MustCompile()`. By default, `regexache` caches entries. |
| REGEXACHE_OUTPUT | File to output the cache contents to. Default: Empty (Don't output cache). |
| REGEXACHE_OUTPUT_MIN | Minimum number of lookups entries need to include when listing cache entries. Default: 1. |
| REGEXACHE_OUTPUT_INTERVAL | If outputing the cache, output every X milliseconds. Default: 1000 (1 second). |
| REGEXACHE_STANDARDIZE| Standardize expressions before caching. Default: Empty (Don't standardize). |
| REGEXACHE_LRU_SIZE | LRU cache size. Default: 1000 |

## Tests

Control (not using the cache).
<br/>**Results** - Single VPC: 6.76GB, Two AppRunner: 17.89GB

```
export REGEXACHE_OFF=1
```

Example of a running memory profile test of a single VPC acceptance test:

```
TF_ACC=1 go test \
    ./internal/service/ec2/... \
    -v -parallel 1 \
    -run='^TestAccVPC_basic$' \
    -cpuprofile cpu.prof \
    -memprofile mem.prof \
    -bench \
    -timeout 60m
pprof -http=localhost:4599 mem.prof
```

Example of a running memory profile test of two parallel AppRunner acceptance tests:

```
TF_ACC=1 go test \
    ./internal/service/apprunner/... \
    -v -parallel 2 \
    -run='TestAccAppRunnerService_ImageRepository_autoScaling|TestAccAppRunnerService_ImageRepository_basic' \
    -cpuprofile cpu.prof \
    -memprofile mem.prof \
    -bench \
    -timeout 60m
pprof -http=localhost:4599 mem.prof
```
