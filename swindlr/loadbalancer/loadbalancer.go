package loadbalancer

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type HealthStatus struct {
	URL     string
	Alive   bool
	Latency time.Duration
}

type contextKey int

const (
	AttemptsKey contextKey = iota
	RetryKey
)

var HealthUpdates = make(chan HealthStatus)

func GetAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(AttemptsKey).(int); ok {
		return attempts
	}
	return 1
}

func GetRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(RetryKey).(int); ok {
		return retry
	}
	return 1
}

func CreateReverseProxy(serverURL *url.URL, sp *ServerPool) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(serverURL)
	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
		log.Printf("[%s %s\n]", serverURL.Host, e.Error())
		retries := GetRetryFromContext(request)

		if retries < 3 {
			time.Sleep(10 * time.Millisecond)
			ctx := context.WithValue(request.Context(), RetryKey, retries+1)
			proxy.ServeHTTP(writer, request.WithContext(ctx))
			return
		}

		sp.MarkBackendStatus(serverURL, false)

		attempts := GetAttemptsFromContext(request)
		log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
		ctx := context.WithValue(request.Context(), AttemptsKey, attempts+1)
		LB(writer, request.WithContext(ctx), sp)
	}
	return proxy
}

func LB(w http.ResponseWriter, r *http.Request, sp *ServerPool) {
	useStickySessions := viper.GetBool("use_sticky_sessions")

	if useStickySessions {
		sessionID, err := r.Cookie("SESSION_ID")
		if err != nil || sessionID == nil {
			newID := generateSessionID()
			sessionID = &http.Cookie{
				Name:     "SESSION_ID",
				Value:    newID,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			}

			http.SetCookie(w, sessionID)
			log.Printf("New session created: %s", newID)
		}
	}

	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer := sp.GetNextPeer(r)

	if peer == nil {
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer.IncrementConnections()
	defer peer.DecrementConnections()

	rateLimitedProxy := RateLimitMiddleware(peer.ReverseProxy, peer)
	rateLimitedProxy.ServeHTTP(w, r)
}

func BackendStatus(u *url.URL) (bool, time.Duration) {
	timeout := 2 * time.Second
	start := time.Now()
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	latency := time.Since(start)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false, latency
	}
	defer conn.Close()
	return true, latency
}

func ManageHealthUpdate() {
	for status := range HealthUpdates {
		fmt.Printf("Received health update for %s: %t\n", status.URL, status.Alive)
		//alerts, metrics, health updates
	}
}

func Health(sp *ServerPool) {
	t := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-t.C:
			sp.HealthCheck(HealthUpdates)
		}
	}
}

func generateSessionID() string {
	return uuid.New().String()
}
