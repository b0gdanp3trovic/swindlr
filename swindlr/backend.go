package main

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
	Connections  int
}

func (b *Backend) setAlive(alive bool) {
	b.mux.Lock()
	b.Alive = alive
	b.mux.Unlock()
}

func (b *Backend) isAlive() (alive bool) {
	b.mux.RLock()
	alive = b.Alive
	b.mux.RUnlock()
	return
}

func (b *Backend) IncrementConnections() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Connections++
}

func (b *Backend) DecrementConnections() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Connections--
}
