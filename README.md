# regexache
Testing

## Environment Variables

| Env Var | Description |
| --- | --- |
| REGEXACHE_OFF | Any value will turn `regexache` completely off. Useful for testing with and without caching. When off, `regexache.MustCompile()` is equivalent to `regexp.MustCompile()`. By default, `regexache` caches entries. |
| REGEXACHE_MAINTENANCE_INTERVAL | Milliseconds between cache maintenance cycles. Default: 30000 (30 seconds). |
| REGEXACHE_EXPIRATION | Milliseconds until an entry expires and `regexache` can remove it from the cache. If and when an entry is removed also depends on `REGEXACHE_MAINTENANCE_INTERVAL` and `REGEXACHE_MINIMUM_USES`. Default: 10000 (10 seconds). |
| REGEXACHE_MINIMUM_USES | If you look up an entry more than this number of times, `regexache` will not remove it from the cache regardless of `REGEXACHE_EXPIRATION`. A value of `0` means that the number of times you lookup an entry is disregarded and `regexache` will remove the entry for expiration. |
| REGEXACHE_CLEAN_TIME | Milliseconds to spend cleaning the cache each maintenance cycle. The cache is locked during cleaning so longer times may reduce performance. Default: 1000 (1 second). |
