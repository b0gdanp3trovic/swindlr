package loadbalancer

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestRateLimitingMiddleware(t *testing.T) {
	viper.Set("rate_limiting.rate", 1)
	viper.Set("rate_limiting.bucket_size", 1)

	// Create a backend with these settings
	serverURL, _ := url.Parse("http://example.com")
	serverPool := NewServerPool(nil)
	backend := CreateNewBackend(serverURL, serverPool)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testServer := httptest.NewServer(RateLimitMiddleware(testHandler, backend))
	defer testServer.Close()

	client := &http.Client{}

	resp, err := client.Get(testServer.URL)
	if err != nil {
		t.Fatalf("Failed to send request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	resp, err = client.Get(testServer.URL)
	if err != nil {
		t.Fatalf("Failed to send request: %s", err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status code %d for rate limit exceeded, got %d", http.StatusTooManyRequests, resp.StatusCode)
	}

	time.Sleep(1 * time.Second)

	resp, err = client.Get(testServer.URL)
	if err != nil {
		t.Fatalf("Failed to send request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d after rate limit reset", http.StatusOK, resp.StatusCode)
	}
}
