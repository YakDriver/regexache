// Package regexache provides a thread-safe regular expression cache.
package regexache

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// nolint
const (
	outputMinDefault      = 1
	outputIntervalDefault = time.Millisecond * 1000

	REGEXACHE_OFF             = "REGEXACHE_OFF"
	REGEXACHE_OUTPUT          = "REGEXACHE_OUTPUT"
	REGEXACHE_OUTPUT_INTERVAL = "REGEXACHE_OUTPUT_INTERVAL"
	REGEXACHE_OUTPUT_MIN      = "REGEXACHE_OUTPUT_MIN"
	REGEXACHE_PRELOAD_OFF     = "REGEXACHE_PRELOAD_OFF"
	REGEXACHE_STANDARDIZE     = "REGEXACHE_STANDARDIZE"
)

//go:embed preload.txt
var preload string

var (
	// mu guards lookups.
	mu sync.RWMutex

	// caching is read on every MustCompile call. An atomic avoids the
	// reader-count contention an RWMutex causes under parallel use.
	caching atomic.Bool

	outputMin      int64
	outputFile     string
	outputInterval time.Duration
	standardizing  bool
)

func init() {
	lookups = make(map[string]int)

	caching.Store(os.Getenv(REGEXACHE_OFF) == "")

	outputInterval = outputIntervalDefault
	if v := os.Getenv(REGEXACHE_OUTPUT_INTERVAL); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		outputInterval = time.Millisecond * time.Duration(i)
	}

	outputMin = outputMinDefault
	if v := os.Getenv(REGEXACHE_OUTPUT_MIN); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		outputMin = int64(i)
	}

	if v := os.Getenv(REGEXACHE_OUTPUT); v != "" && caching.Load() {
		outputFile = v
		go func() {
			for {
				time.Sleep(outputInterval)
				outputCache()
			}
		}()
	}

	// Set before preloading so preloaded keys match what MustCompile looks up.
	standardizing = os.Getenv(REGEXACHE_STANDARDIZE) != ""

	if v := os.Getenv(REGEXACHE_PRELOAD_OFF); v == "" && caching.Load() {
		preloadPatterns(preload)
	}
}

// preloadPatterns compiles and caches each newline-separated pattern, keyed
// the same way MustCompile looks it up.
func preloadPatterns(patterns string) {
	for r := range strings.SplitSeq(patterns, "\n") {
		if r == "" {
			continue
		}

		if standardizing {
			r = standardize(r)
		}

		cache.Store(r, regexp.MustCompile(r))
	}
}

var cache sync.Map
var lookups map[string]int

// MustCompile is like regexp.MustCompile but caches compiled regular expressions.
func MustCompile(str string) *regexp.Regexp {
	if !caching.Load() {
		return regexp.MustCompile(str)
	}

	if standardizing {
		str = standardize(str)
	}

	if outputFile != "" {
		mu.Lock()
		lookups[str]++
		mu.Unlock()
	}

	if re, ok := cache.Load(str); ok {
		return re.(*regexp.Regexp)
	}

	re, _ := cache.LoadOrStore(str, regexp.MustCompile(str))
	return re.(*regexp.Regexp)
}

// SetCaching enables or disables regex caching.
func SetCaching(enabled bool) {
	caching.Store(enabled)
}

// IsCachingEnabled returns whether regex caching is currently enabled.
func IsCachingEnabled() bool {
	return caching.Load()
}

func outputCache() {
	filename := fmt.Sprintf("%s.%s", outputFile, strings.ReplaceAll(uuid.New().String(), "-", ""))

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return // Fail silently for output functionality
	}
	defer func() { _ = f.Close() }()

	if _, err = f.WriteString("regex\tcount\n"); err != nil {
		return
	}

	cache.Range(func(k, _ any) bool {
		pattern := k.(string)

		mu.RLock()
		count := lookups[pattern]
		mu.RUnlock()

		if count >= int(outputMin) {
			_, _ = fmt.Fprintf(f, "%s\t%d\n", pattern, count)
		}
		return true
	})
}
