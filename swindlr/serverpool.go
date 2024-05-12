package main

import (
	"fmt"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
)

type ServerPool struct {
	backends []*Backend
	current  uint64
	mux      sync.RWMutex
}

func (s *ServerPool) AddBackend(backend *Backend) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.backends = append(s.backends, backend)
	log.Printf("Added a new backend with url %s", backend.URL)
}

func (s *ServerPool) RemoveBackend(URL string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	for i, backend := range s.backends {
		if backend.URL.String() == URL {
			s.backends = append(s.backends[:i], s.backends[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("backend not found with url URL %s", URL)
}

func (s *ServerPool) NextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *ServerPool) GetNextPeer() *Backend {
	next := s.NextIndex()
	l := len(s.backends) + next
	for i := next; i < l; i++ {
		idx := i % len(s.backends)
		if s.backends[idx].isAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}
	return nil
}

func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.URL.String() == backendUrl.String() {
			b.setAlive(alive)
			break
		}
	}
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		alive := isBackendAlive(b.URL)
		b.setAlive(alive)
		healthUpdates <- HealthStatus{URL: b.URL.String(), Alive: alive}
	}
}
