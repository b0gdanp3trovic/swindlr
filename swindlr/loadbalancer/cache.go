package loadbalancer

import (
	"sync"
	"time"
)

type CacheItem struct {
	content    []byte
	expiration time.Time
}

type Cache struct {
	items map[string]CacheItem
	mux   sync.RWMutex
	ttl   time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[string]CacheItem),
		ttl:   ttl,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mux.Lock()
	defer c.mux.RUnlock()
	item, found := c.items[key]
	if !found || time.Now().After(item.expiration) {
		return nil, false
	}

	return item.content, true
}

func (c *Cache) Set(key string, content []byte, duration time.Duration) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.items[key] = CacheItem{
		content:    content,
		expiration: time.Now().Add(duration),
	}
}

func (c *Cache) DeleteExpired() {
	c.mux.Lock()
	defer c.mux.Unlock()
	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiration) {
			delete(c.items, key)
		}
	}
}
