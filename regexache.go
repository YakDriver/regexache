package regexache

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"sync"
	"time"

	"golang.org/x/exp/slices"
)

const (
	maintenanceIntervalDefault = time.Duration(0)
	expirationDefault          = time.Millisecond * 10000
	minimumUsesDefault         = int64(2)
	cleanTimeDefault           = time.Millisecond * 1000
	outputMinDefault           = 10

	REGEXACHE_OFF                  = "REGEXACHE_OFF"
	REGEXACHE_MAINTENANCE_INTERVAL = "REGEXACHE_MAINTENANCE_INTERVAL"
	REGEXACHE_EXPIRATION           = "REGEXACHE_EXPIRATION"
	REGEXACHE_MINIMUM_USES         = "REGEXACHE_MINIMUM_USES"
	REGEXACHE_CLEAN_TIME           = "REGEXACHE_CLEAN_TIME"
	REGEXACHE_CACHE_OUTPUT         = "REGEXACHE_CACHE_OUTPUT"
	REGEXACHE_CACHE_OUTPUT_MIN     = "REGEXACHE_CACHE_OUTPUT_MIN"
)

var (
	mutex *sync.RWMutex
	once  *sync.Once

	caching             bool
	maintainCache       bool
	maintenanceInterval time.Duration
	expiration          time.Duration
	minimumUses         int64
	cleanTime           time.Duration
	outputMin           int64
	outputFile          string
	bobCentry           *centry
)

func init() {
	mutex = &sync.RWMutex{}
	cache = make(map[string]centry)
	once = &sync.Once{}

	caching = true
	if v := os.Getenv(REGEXACHE_OFF); v != "" {
		caching = false
	}

	maintainCache = true
	maintenanceInterval = maintenanceIntervalDefault
	if v := os.Getenv(REGEXACHE_MAINTENANCE_INTERVAL); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		if i < 1 {
			maintainCache = false
		} else {
			maintenanceInterval = time.Millisecond * time.Duration(i)
		}
	}

	expiration = expirationDefault
	if v := os.Getenv(REGEXACHE_EXPIRATION); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		expiration = time.Millisecond * time.Duration(i)
	}

	minimumUses = minimumUsesDefault
	if v := os.Getenv(REGEXACHE_MINIMUM_USES); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		minimumUses = int64(i)
	}

	cleanTime = cleanTimeDefault
	if v := os.Getenv(REGEXACHE_CLEAN_TIME); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		cleanTime = time.Millisecond * time.Duration(i)
	}

	if v := os.Getenv(REGEXACHE_CACHE_OUTPUT); v != "" {
		bobCentry = &centry{}
		outputFile = v
		go func() {
			runtime.SetFinalizer(bobCentry, func(_ *centry) {
				outputCache()
			})
		}()
	}

	outputMin = outputMinDefault
	if v := os.Getenv(REGEXACHE_CACHE_OUTPUT_MIN); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}

		outputMin = int64(i)
	}
}

type centry struct {
	re      *regexp.Regexp
	count   int64
	lastUse int64
}

var cache map[string]centry

func clean() {
	mutex.Lock()
	defer mutex.Unlock()

	endCleanTime := time.Now().Add(cleanTime)
	for k, v := range cache {
		if (v.count < minimumUses || minimumUses == 0) && (time.Now().UnixNano()-v.lastUse) > int64(expiration) {
			delete(cache, k)
		}

		if time.Now().After(endCleanTime) {
			break
		}
	}
}

func maintain() {
	if !maintainCache {
		return
	}

	for {
		time.Sleep(maintenanceInterval)
		clean()
	}
}

func MustCompile(str string) *regexp.Regexp {
	if !caching {
		return regexp.MustCompile(str)
	}

	once.Do(func() {
		go maintain()
	})

	if v := lookup(str); v != nil {
		return v
	}

	mutex.Lock()
	defer mutex.Unlock()

	cache[str] = centry{
		re:      regexp.MustCompile(str),
		count:   1,
		lastUse: time.Now().UnixNano(),
	}

	return cache[str].re
}

func lookup(str string) *regexp.Regexp {
	mutex.RLock()
	defer mutex.RUnlock()

	if v, ok := cache[str]; ok {
		v.count++
		v.lastUse = time.Now().UnixNano()
		return v.re
	}

	return nil
}

func outputCache() {
	var uses []int64
	for _, v := range cache {
		uses = append(uses, v.count)
	}

	slices.Sort(uses)
	slices.Reverse(uses)

	f, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	_, err = f.WriteString("regex\tcount\n")
	if err != nil {
		panic(err)
	}

	for _, v := range uses {
		if v < outputMin {
			continue
		}

		for k, c := range cache {
			if c.count == v {
				_, err := f.WriteString(fmt.Sprintf("%s\t%d\n", k, v))
				if err != nil {
					panic(err)
				}
			}
		}
	}
}
