package loadbalancer

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/spf13/viper"
	"golang.org/x/time/rate"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
	Connections  int
	SessionMap   map[string]*Backend
	Limiter      *rate.Limiter
}

func (b *Backend) setAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) IncrementConnections() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Connections++
}

func (b *Backend) DecrementConnections() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Connections--
}

func CreateNewBackend(serverURL *url.URL, serverPool *ServerPool) *Backend {
	r := float64(viper.GetInt("rate_limiting.rate"))
	bucketSize := viper.GetInt("rate_limiting.bucket_size")

	limiter := rate.NewLimiter(rate.Limit(r), bucketSize)
	return &Backend{
		URL:          serverURL,
		Alive:        true,
		ReverseProxy: CreateReverseProxy(serverURL, serverPool),
		Limiter:      limiter,
	}
}

func RateLimitMiddleware(next http.Handler, backend *Backend) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !backend.Limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
