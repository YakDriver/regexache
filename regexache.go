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
	mutex *sync.RWMutex

	caching        bool
	outputMin      int64
	outputFile     string
	outputInterval time.Duration
	standardizing  bool
)

func init() {
	mutex = &sync.RWMutex{}
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
	if !caching {
		return regexp.MustCompile(str)
	}

	if standardizing {
		str = standardize(str)
	}

	if outputFile != "" {
		mutex.Lock()
		if _, ok := lookups[str]; !ok {
			lookups[str] = 0
		}
		lookups[str]++
		mutex.Unlock()
	}

	re, ok := cache.Load(str)
	if ok {
		return re.(*regexp.Regexp)
	}

	re, _ = cache.LoadOrStore(str, regexp.MustCompile(str))
	return re.(*regexp.Regexp)
}

// SetCaching enables or disables regex caching.
func SetCaching(enabled bool) {
	caching = enabled
}

// IsCachingEnabled returns whether regex caching is currently enabled.
func IsCachingEnabled() bool {
	return caching
}

func outputCache() {
	filename := fmt.Sprintf("%s.%s", outputFile, strings.Replace(uuid.New().String(), "-", "", -1))

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	_, err = f.WriteString("regex\tcount\n")
	if err != nil {
		panic(err)
	}

	cache.Range(func(k any, _ any) bool {
		mutex.RLock()
		v, ok := lookups[k.(string)]
		if !ok {
			v = 0
		}
		mutex.RUnlock()

		if v < int(outputMin) {
			return true
		}

		_, err := f.WriteString(fmt.Sprintf("%s\t%d\n", k.(string), v))
		if err != nil {
			panic(err)
		}
		return true
	})
}
