package loadbalancer

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/spf13/viper"
)

func TestGetBackendBySessionID(t *testing.T) {
	sp := NewServerPool(nil)
	backend1 := &Backend{
		URL:   parseURL("http://backend1.test"),
		Alive: true,
	}

	backend2 := &Backend{
		URL:   parseURL("http://backend2.test"),
		Alive: true,
	}

	sp.backends = []*Backend{backend1, backend2}
	sessionID := "session123"
	sp.sessions[sessionID] = backend1

	result := sp.GetBackendBySessionID(sessionID)

	if result != backend1 {
		t.Errorf("Expected backend1, got %v", result)
	}
}

func parseURL(urlStr string) *url.URL {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}
	return parsedURL
}

func TestRemoveBackend(t *testing.T) {
	sp := NewServerPool(nil)
	url1 := parseURL("http://backend1.test")
	url2 := parseURL("http://backend2.test")

	backend1 := &Backend{
		URL:   url1,
		Alive: true,
	}

	backend2 := &Backend{
		URL:   url2,
		Alive: true,
	}

	sp.AddBackend(backend1)
	sp.AddBackend(backend2)

	err := sp.RemoveBackend("http://backend1.test")
	if err != nil {
		t.Errorf("Failed to remove existing backend: %v", err)
	}

	if len(sp.backends) != 1 || sp.backends[0] != backend2 {
		t.Errorf("Backend1 was not removed correctly.")
	}

	err = sp.RemoveBackend("http://nonexistent.test")
	if err == nil {
		t.Error("Expected an error while trying to remove a backend that does not exist.")
	}
}

func setupServerPool(algorithm Algorithm) *ServerPool {
	serverPool := NewServerPool(algorithm)
	serverUrl, _ := url.Parse("http://backend1.test")
	backend1 := CreateNewBackend(serverUrl, serverPool)
	serverPool.AddBackend(backend1)
	return serverPool
}

func TestStickySessions(t *testing.T) {
	req := httptest.NewRequest("GET", "http://backend.test", nil)

	t.Run("Sticky Sessions enabled with valid session", func(t *testing.T) {
		viper.Set("use_sticky_sessions", true)
		sp := setupServerPool(&RoundRobin{})
		sp.sessions["session123"] = sp.backends[0]

		req.AddCookie(&http.Cookie{Name: "SESSION_ID", Value: "session123"})

		backend := sp.GetNextPeer(req)
		if backend != sp.backends[0] {
			t.Errorf("Expected backend %v, got %v", sp.backends[0], backend)
		}
	})

	t.Run("Sticky sessions enabled without valid session", func(t *testing.T) {
		viper.Set("use_sticky_sessions", true)
		sp := setupServerPool(&RoundRobin{})

		req.Header.Set("Cookie", "")

		// If no sessions are available, the next peer
		// should be determined algoritmically
		backend := sp.GetNextPeer(req)

		if backend == nil {
			t.Errorf("Expected a backend, got nil")
		}
	})

	t.Run("Sticky sessions disabled", func(t *testing.T) {
		viper.Set("use_sticky_sessions", false)
		sp := setupServerPool(&RoundRobin{})

		backend := sp.GetNextPeer(req)
		if backend == nil {
			t.Errorf("Expected a backend, got nil")
		}
	})

	t.Run("No backends available", func(t *testing.T) {
		sp := NewServerPool(&RoundRobin{})

		backend := sp.GetNextPeer(req)
		if backend != nil {
			t.Errorf("Expected no backend, got %v", backend)
		}
	})
}
