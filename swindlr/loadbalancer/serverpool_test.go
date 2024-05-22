package loadbalancer

import (
	"net/url"
	"testing"
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
