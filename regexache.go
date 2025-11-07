package regexache

import (
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
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
	mu sync.RWMutex

	caching        bool
	outputMin      int64
	outputFile     string
	outputInterval time.Duration
	standardizing  bool
)

func init() {
	lookups = make(map[string]int)

	caching = true
	if v := os.Getenv(REGEXACHE_OFF); v != "" {
		caching = false
	}

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

	if v := os.Getenv(REGEXACHE_OUTPUT); v != "" && caching {
		outputFile = v
		go func() {
			for {
				time.Sleep(outputInterval)
				outputCache()
			}
		}()
	}

	if v := os.Getenv(REGEXACHE_PRELOAD_OFF); v == "" && caching {
		preRE := strings.Split(preload, "\n")
		for _, r := range preRE {
			if r == "" {
				continue
			}

			cache.Store(r, regexp.MustCompile(r))
		}
	}

	standardizing = false
	if v := os.Getenv(REGEXACHE_STANDARDIZE); v != "" {
		standardizing = true
	}
}

var cache sync.Map
var lookups map[string]int

func MustCompile(str string) *regexp.Regexp {
	mu.RLock()
	cachingEnabled := caching
	mu.RUnlock()

	if !cachingEnabled {
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
	mu.Lock()
	caching = enabled
	mu.Unlock()
}

// IsCachingEnabled returns whether regex caching is currently enabled.
func IsCachingEnabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return caching
}

func outputCache() {
	filename := fmt.Sprintf("%s.%s", outputFile, strings.ReplaceAll(uuid.New().String(), "-", ""))

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return // Fail silently for output functionality
	}
	defer f.Close()

	if _, err = f.WriteString("regex\tcount\n"); err != nil {
		return
	}

	cache.Range(func(k, _ any) bool {
		pattern := k.(string)
		
		mu.RLock()
		count := lookups[pattern]
		mu.RUnlock()

		if count >= int(outputMin) {
			f.WriteString(fmt.Sprintf("%s\t%d\n", pattern, count))
		}
		return true
	})
}
