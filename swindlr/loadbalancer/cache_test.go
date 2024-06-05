package loadbalancer

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestCacheMiddleware(t *testing.T) {
	viper.Set("use_cache", true)

	cache := NewCache(1 * time.Minute)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", "12345")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	cacheHandler := CacheMiddleware(cache, handler)

	req, _ := http.NewRequest("GET", "/test", nil)

	rr := httptest.NewRecorder()
	cacheHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "12345", rr.Header().Get("ETag"))
	assert.Equal(t, "Hello, World!", rr.Body.String())
	assert.Equal(t, "MISS", rr.Header().Get("X-Swindlr-Cache"))

	rr = httptest.NewRecorder()
	cacheHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "12345", rr.Header().Get("ETag"))
	assert.Equal(t, "Hello, World!", rr.Body.String())
	assert.Equal(t, "HIT", rr.Header().Get("X-Swindlr-Cache"))
}

func TestCacheExpiration(t *testing.T) {
	viper.Set("use_cache", true)

	cache := NewCache(1 * time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", "12345")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	cacheHandler := CacheMiddleware(cache, handler)

	req, _ := http.NewRequest("GET", "/test", nil)

	rr := httptest.NewRecorder()
	cacheHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "12345", rr.Header().Get("ETag"))
	assert.Equal(t, "Hello, World!", rr.Body.String())
	assert.Equal(t, "MISS", rr.Header().Get("X-Swindlr-Cache"))

	time.Sleep(2 * time.Second)

	rr = httptest.NewRecorder()
	cacheHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "12345", rr.Header().Get("ETag"))
	assert.Equal(t, "Hello, World!", rr.Body.String())
	assert.Equal(t, "MISS", rr.Header().Get("X-Swindlr-Cache"))
}
