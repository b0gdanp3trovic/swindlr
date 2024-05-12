package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type HealthStatus struct {
	URL   string
	Alive bool
}

type contextKey int

const (
	AttemptsKey contextKey = iota
	RetryKey
)

var healthUpdates = make(chan HealthStatus)

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

func CreateReverseProxy(serverURL *url.URL) *httputil.ReverseProxy {
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

		serverPool.MarkBackendStatus(serverURL, false)

		attempts := GetAttemptsFromContext(request)
		log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
		ctx := context.WithValue(request.Context(), AttemptsKey, attempts+1)
		lb(writer, request.WithContext(ctx))
	}
	return proxy
}

func CreateNewBackend(serverURL *url.URL) *Backend {
	backend := Backend{
		URL:          serverURL,
		Alive:        true,
		ReverseProxy: CreateReverseProxy(serverURL),
	}

	return &backend
}

func lb(w http.ResponseWriter, r *http.Request) {
	attempts := GetAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
		return
	}

	peer := serverPool.GetNextPeer()
	if peer != nil {
		peer.IncrementConnections()
		defer peer.DecrementConnections()
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func isBackendAlive(u *url.URL) bool {
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", u.Host, timeout)
	if err != nil {
		log.Println("Site unreachable, error: ", err)
		return false
	}
	defer conn.Close()
	return true
}

func manageHealthUpdate() {
	for status := range healthUpdates {
		fmt.Printf("Received health update for %s: %t\n", status.URL, status.Alive)
		//alerts, metrics
	}
}

func healthcheck() {
	t := time.NewTicker(time.Minute * 2)
	for {
		select {
		case <-t.C:
			serverPool.HealthCheck()
		}
	}
}
