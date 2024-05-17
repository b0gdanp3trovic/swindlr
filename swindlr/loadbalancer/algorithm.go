package loadbalancer

import (
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

type Algorithm interface {
	SelectBackend(request *http.Request) *Backend
}

type RoundRobin struct {
	backends []*Backend
	current  uint64
}

func (rr *RoundRobin) SelectBackend(request *http.Request) *Backend {
	if len(rr.backends) == 0 {
		return nil
	}
	index := atomic.AddUint64(&rr.current, 1) % uint64(len(rr.backends))
	return rr.backends[index]
}

type LeastConnections struct {
	backends []*Backend
}

func (lc *LeastConnections) SelectBackend(request *http.Request) *Backend {
	var minBackend *Backend
	minConnections := int(^uint(0) >> 1)

	for _, backend := range lc.backends {
		backend.mux.Lock()
		if backend.Alive && (minBackend == nil || backend.Connections < minConnections) {
			minBackend = backend
			minConnections = backend.Connections
		}
		backend.mux.Unlock()
	}
	return minBackend
}

type Random struct {
	backends []*Backend
	rand     *rand.Rand
}

func NewRandom(backends []*Backend) *Random {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	return &Random{
		backends: backends,
		rand:     r,
	}
}

func (r *Random) SelectBackend(request *http.Request) *Backend {
	if len(r.backends) == 0 {
		return nil
	}

	index := r.rand.Intn(len(r.backends))
	return r.backends[index]
}
