# regexache

`regexache` is a thread-safe regular expression cache, providing a drop-in replacement for `regexp.MustCompile()` (`regexache` calls `regexp.MustCompile()` on your behalf to populate the cache). This special purpose cache specifically addresses regular expressions, which use a lot of memory. In a [project](https://github.com/hashicorp/terraform-provider-aws) with about ~4500 regexes, using `regexache` saved nearly 20% total memory use.

Unlike excellent caches, such as [go-cache](https://github.com/patrickmn/go-cache) or memcached, the calling code does not need to know anything about the cache or instantiate it, simply dropping in `regexache.MustCompile()` in place of `regexp.MustCompile()`. There are cons to this approach but for an existing project, they are outweighed by not needing to rework existing code (other than the drop in).

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
| REGEXACHE_OFF | Any value will turn `regexache` completely off. Useful for testing with and without caching. When off, `regexache.MustCompile()` is equivalent to `regexp.MustCompile()`. By default, `regexache` caches entries. |
| REGEXACHE_MAINTENANCE_INTERVAL | Milliseconds between cache maintenance cycles. A value of `0` means that the cache will not be maintained--no entries will be removed. Default: 0 (entries won't be removed). |
| REGEXACHE_EXPIRATION | Milliseconds until an entry expires and `regexache` can remove it from the cache. If and when an entry is removed also depends on `REGEXACHE_MAINTENANCE_INTERVAL` and `REGEXACHE_MINIMUM_USES`. Default: 10000 (10 seconds). |
| REGEXACHE_MINIMUM_USES | If you look up an entry more than this number of times, `regexache` will not remove it from the cache regardless of `REGEXACHE_EXPIRATION`. A value of `0` means that the number of times you lookup an entry is disregarded and `regexache` will remove the entry for expiration. |
| REGEXACHE_CLEAN_TIME | Milliseconds to spend cleaning the cache each maintenance cycle. The cache is locked during cleaning so longer times may reduce performance. Default: 1000 (1 second). |

## Examples

Don't use cache. (Single VPC: 6.76GB, 2-AppRunner: 17.89GB)

```
export REGEXACHE_OFF=1
```

Single VPC: 5.62GB, 2-AppRunner: 15.01GB

```
export REGEXACHE_MAINTENANCE_INTERVAL=1000
export REGEXACHE_EXPIRATION=500
export REGEXACHE_MINIMUM_USES=10
export REGEXACHE_CLEAN_TIME=100
```

Single VPC: 5.54GB, 2-AppRunner: 14.81GB

```
export REGEXACHE_MAINTENANCE_INTERVAL=0
```

Single VPC: 5.67GB, 2-AppRunner: 15.05GB

```
export REGEXACHE_MAINTENANCE_INTERVAL=2000
export REGEXACHE_EXPIRATION=2000
export REGEXACHE_MINIMUM_USES=3
export REGEXACHE_CLEAN_TIME=100
```

Single VPC: 5.62GB, 2-AppRunner: 15.00GB

```
export REGEXACHE_MAINTENANCE_INTERVAL=5000
export REGEXACHE_EXPIRATION=1000
export REGEXACHE_MINIMUM_USES=2
export REGEXACHE_CLEAN_TIME=500
```

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
