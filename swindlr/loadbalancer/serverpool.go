package loadbalancer

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type ServerPool struct {
	backends  []*Backend
	mux       sync.RWMutex
	algorithm Algorithm
	sessions  map[string]*Backend
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

func (s *ServerPool) GetBackendBySessionID(sessionID string) *Backend {
	s.mux.RLock()
	backend, exists := s.sessions[sessionID]
	s.mux.RUnlock()

	if exists {
		return backend
	}

	return nil
}

func (s *ServerPool) AssignSessionToBackend(sessionID string, backend *Backend) {
	s.mux.Lock()
	s.sessions[sessionID] = backend
	s.mux.Unlock()
}

func (s *ServerPool) GetNextPeer(r *http.Request) *Backend {
	//Check if sessionID exists in the request cookies
	var sessionID *http.Cookie
	var err error

	useStickySessions := viper.GetBool("use_sticky_sessions")
	if useStickySessions {
		sessionID, err = r.Cookie("SESSION_ID")
		if err == nil && sessionID != nil {
			backend := s.GetBackendBySessionID(sessionID.Value)
			if backend != nil {
				return backend
			}
		}
	}

	//There is no valid session, use an algorithm
	//to assign backend and store it
	newBackend := s.algorithm.SelectBackend(s.backends)
	if newBackend == nil {
		return nil
	}

	if useStickySessions && sessionID != nil && sessionID.Value != "" {
		s.AssignSessionToBackend(sessionID.Value, newBackend)
	}

	return newBackend
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
		alive, latency := BackendStatus(b.URL)
		b.setAlive(alive)
		b.setLatency(latency)
		healthUpdates <- HealthStatus{URL: b.URL.String(), Alive: alive}
	}
}

func NewServerPool(algorithm Algorithm) *ServerPool {
	return &ServerPool{
		algorithm: algorithm,
		sessions:  make(map[string]*Backend),
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
	case "latency_aware":
		algo = &LatencyAware{}
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
