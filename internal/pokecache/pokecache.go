package pokecache

import (
	"sync"
	"time"
)

var Cache cache = NewCache(time.Second * 30)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type cache struct {
	cache map[string]cacheEntry
	mutex sync.Mutex
}

func (c cache) Add(key string, val []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = cacheEntry{createdAt: time.Now(), val: val}
}

func (c cache) Get(key string) ([]byte, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	val, ok := c.cache[key]
	if ok {
		return val.val, true
	} else {
		return val.val, false
	}
}

func (c cache) reapLoop(lifetime time.Duration) {

	ticker := time.NewTicker(lifetime)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		for k, v := range c.cache {
			if time.Since(v.createdAt) > lifetime {
				delete(c.cache, k)
			}
		}
	}
}

func NewCache(lifetime time.Duration) cache {
	var cache cache
	cache.cache = make(map[string]cacheEntry)

	go cache.reapLoop(lifetime)

	return cache
}
