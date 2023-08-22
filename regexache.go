package regexache

import (
	"regexp"
	"sync"
	"time"
)

const MaintenanceInterval = time.Second * 30
const ProtectedInterval = time.Second * 10
const MinimumUses = 2

var mutex *sync.RWMutex
var once *sync.Once

func init() {
	mutex = &sync.RWMutex{}
	cache = make(map[string]centry)
	once = &sync.Once{}
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

	for k, v := range cache {
		if v.count < MinimumUses && (time.Now().UnixNano()-v.lastUse) > int64(ProtectedInterval) {
			delete(cache, k)
		}
	}
}

func maintain() {
	for {
		time.Sleep(MaintenanceInterval)
		clean()
	}
}

func MustCompile(str string) *regexp.Regexp {
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
