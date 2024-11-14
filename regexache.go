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
	lru "github.com/hashicorp/golang-lru"
)

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
	cache          *lru.Cache
	lookups        map[string]int
	lookupsLock    sync.RWMutex
	caching        bool
	outputMin      int64
	outputFile     string
	outputInterval time.Duration
	standardizing  bool
)

func init() {
	var err error
	cache, err = lru.New(1000)
	if err != nil {
		panic(err)
	}

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

			cache.Add(r, regexp.MustCompile(r))
		}
	}

	standardizing = false
	if v := os.Getenv(REGEXACHE_STANDARDIZE); v != "" {
		standardizing = true
	}
}

func MustCompile(str string) *regexp.Regexp {
	if !caching {
		return regexp.MustCompile(str)
	}

	if standardizing {
		str = standardize(str)
	}

	if outputFile != "" {
		lookupsLock.Lock()
		if lookups == nil {
			lookups = make(map[string]int)
		}
		if _, ok := lookups[str]; !ok {
			lookups[str] = 0
		}
		lookups[str]++
		lookupsLock.Unlock()
	}

	if re, ok := cache.Get(str); ok {
		return re.(*regexp.Regexp)
	}

	re := regexp.MustCompile(str)
	cache.Add(str, re)
	return re
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

	cacheKeys := cache.Keys()
	for _, k := range cacheKeys {
		lookupsLock.RLock()
		v, ok := lookups[k.(string)]
		if !ok {
			v = 0
		}
		lookupsLock.RUnlock()

		if v < int(outputMin) {
			continue
		}

		_, err := f.WriteString(fmt.Sprintf("%s\t%d\n", k.(string), v))
		if err != nil {
			panic(err)
		}
	}
}
