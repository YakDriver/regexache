# regexache
Testing

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
