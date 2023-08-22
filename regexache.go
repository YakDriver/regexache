package regexache

import (
	"regexp"
	"sync"
)

const DefaultCacheSize = 200

var mutex *sync.RWMutex

func init() {
	mutex = &sync.RWMutex{}
	cache = make(map[string]centry)
}

type centry struct {
	re    *regexp.Regexp
	count int
}

var cache map[string]centry

func MustCompile(str string) *regexp.Regexp {
	if v := lookup(str); v != nil {
		return v
	}

	mutex.Lock()
	defer mutex.Unlock()

	cache[str] = centry{
		re:    regexp.MustCompile(str),
		count: 1,
	}

	return cache[str].re
}

func lookup(str string) *regexp.Regexp {
	mutex.RLock()
	defer mutex.RUnlock()

	if v, ok := cache[str]; ok {
		v.count++
		return v.re
	}

	return nil
}
