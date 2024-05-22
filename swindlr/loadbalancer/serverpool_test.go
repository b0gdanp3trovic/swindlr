package loadbalancer

import (
	"net/url"
	"testing"
)

func TestGetBackendBySessionID(t *testing.T) {
	sp := NewServerPool(&RoundRobin{})
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
