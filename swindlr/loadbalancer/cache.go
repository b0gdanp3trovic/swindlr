package loadbalancer

import (
	"bytes"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type CacheItem struct {
	Content      []byte
	Expiration   time.Time
	ETag         string
	LastModified time.Time
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

func (c *Cache) Get(key string) (CacheItem, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	item, found := c.items[key]
	if !found || time.Now().After(item.Expiration) {
		return CacheItem{}, false
	}
	return item, true
}

func (c *Cache) Set(key string, content []byte, headers http.Header, duration time.Duration) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.items[key] = CacheItem{
		Content:      content,
		Expiration:   time.Now().Add(duration),
		ETag:         headers.Get("ETag"),
		LastModified: time.Now(),
	}
}

func (c *Cache) DeleteExpired() {
	c.mux.Lock()
	defer c.mux.Unlock()
	now := time.Now()
	for key, item := range c.items {
		if now.After(item.Expiration) {
			delete(c.items, key)
		}
	}
}

// Middleware logic
type responseWriter struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
	header http.Header
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Header() http.Header {
	return rw.header
}

func CacheMiddleware(cache *Cache, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !viper.GetBool("use_cache") {
			next.ServeHTTP(w, r)
			return
		}

		if item, found := cache.Get(r.URL.Path); found {
			/*
				The If-None-Match HTTP request header makes the request conditional.
				For GET and HEAD methods, the server will return the requested resource, with a 200 status,
				only if it doesn't have an ETag matching the given ones.
			*/
			if match := r.Header.Get("If-None-Match"); match != "" && match == item.ETag {
				w.WriteHeader(http.StatusNotModified)
			}

			/*
				The If-Modified-Since request HTTP header makes the request conditional:
				the server sends back the requested resource, with a 200 status,
				only if it has been last modified after the given date.
			*/

			if modifiedSince := r.Header.Get("If-Modified-Since"); modifiedSince != "" {
				t, err := time.Parse(http.TimeFormat, modifiedSince)
				if err == nil && item.LastModified.Before(t.Add(1*time.Second)) {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}

			w.Header().Set("ETag", item.ETag)
			w.Header().Set("Last-Modified", item.LastModified.Format(http.TimeFormat))
			w.Header().Set("X-Swindlr-Cache", "HIT")
			w.Write(item.Content)
			return
		}

		w.Header().Set("X-Swindlr-Cache", "MISS")
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK, header: make(http.Header)}
		next.ServeHTTP(rw, r)

		if rw.status == http.StatusOK {
			cacheControl := rw.header.Get("Cache-Control")
			if cacheControl != "no-store" && cacheControl != "private" {
				cache.Set(r.URL.Path, rw.body.Bytes(), rw.header, cache.ttl)
			}
		}
	})
}
