package loadbalancer

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"sync"
	"time"
)

type ServerPool struct {
	backends  []*Backend
	mux       sync.RWMutex
	algorithm Algorithm
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

func (s *ServerPool) GetNextPeer() *Backend {
	if s.algorithm == nil {
		log.Fatal("No load balancing algorithm specified")
	}
	return s.algorithm.SelectBackend(s.backends)
}

func (s *ServerPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range s.backends {
		if b.URL.String() == backendUrl.String() {
			log.Printf("Marking %s alive status as %t", backendUrl, alive)
			b.setAlive(alive)
			break
		}
	}
}

func (s *ServerPool) HealthCheck(healthUpdates chan<- HealthStatus) {
	for _, b := range s.backends {
		alive := IsBackendAlive(b.URL)
		b.setAlive(alive)
		healthUpdates <- HealthStatus{URL: b.URL.String(), Alive: alive}
	}
}

func NewServerPool(algorithm Algorithm) *ServerPool {
	return &ServerPool{
		algorithm: algorithm,
	}
}

func SetupServerPool(backendURLs []string, strategy string) *ServerPool {
	var algo Algorithm
	switch strategy {
	case "round_robin":
		algo = &RoundRobin{}
	case "least_connections":
		algo = &LeastConnections{}
	case "random":
		randSrc := rand.NewSource(time.Now().UnixNano())
		algo = &Random{rand: rand.New(randSrc)}
	default:
		log.Fatalf("Unknown load balancing strategy: %s", strategy)
	}

	serverPool := NewServerPool(algo)

	for _, urlStr := range backendURLs {
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			log.Fatalf("Error parsing backend URL: %s", err)
		}
		backend := CreateNewBackend(parsedURL, serverPool)
		serverPool.AddBackend(backend)
	}

	return serverPool
}
