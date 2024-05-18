package loadbalancer

import (
	"math/rand"
	"sync/atomic"
	"time"
)

type Algorithm interface {
	SelectBackend(backends []*Backend) *Backend
}

type RoundRobin struct {
	current uint64
}

func (rr *RoundRobin) SelectBackend(backends []*Backend) *Backend {
	if len(backends) == 0 {
		return nil
	}
	index := atomic.AddUint64(&rr.current, 1) % uint64(len(backends))
	return backends[index]
}

type LeastConnections struct{}

func (lc *LeastConnections) SelectBackend(backends []*Backend) *Backend {
	var minBackend *Backend
	minConnections := int(^uint(0) >> 1)

	for _, backend := range backends {
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

func NewRandom() *Random {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	return &Random{
		rand: r,
	}
}

func (r *Random) SelectBackend(backends []*Backend) *Backend {
	if len(backends) == 0 {
		return nil
	}

	index := r.rand.Intn(len(r.backends))
	return r.backends[index]
}
