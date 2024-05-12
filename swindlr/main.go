package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	Attempts int = iota
	Retry
)

var serverPool ServerPool

func main() {
	var customPath string
	flag.StringVar(&customPath, "configPath", "", "Custom path to the config directory")
	flag.Parse()

	initConfig(customPath)

	port := viper.GetInt("port")
	backendURLs := viper.GetStringSlice("backends")
	useSSL := viper.GetBool("use_ssl")
	certPath := viper.GetString("ssl_cert_file")
	keyPath := viper.GetString("ssl_key_file")

	for _, tok := range backendURLs {
		serverUrl, err := url.Parse(tok)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, e error) {
			log.Printf("[%s %s\n]", serverUrl.Host, e.Error())
			retries := GetRetryFromContext(request)

			if retries < 3 {
				time.Sleep(10 * time.Millisecond)
				ctx := context.WithValue(request.Context(), Retry, retries+1)
				proxy.ServeHTTP(writer, request.WithContext(ctx))
				return
			}

			serverPool.MarkBackendStatus(serverUrl, false)

			attempts := GetAttemptsFromContext(request)
			log.Printf("%s(%s) Attempting retry %d\n", request.RemoteAddr, request.URL.Path, attempts)
			ctx := context.WithValue(request.Context(), Attempts, attempts+1)
			lb(writer, request.WithContext(ctx))
		}

		serverPool.AddBackend(&Backend{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: proxy,
		})

		log.Printf("Configured server: %s\n", serverUrl)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(lb),
	}

	go healthcheck()
	go manageHealthUpdate()

	// Prepare API endpoints
	gin.SetMode(gin.ReleaseMode)
	apiRouter := gin.Default()
	apiRouter.POST("/api/backends", AddBackend)
	apiRouter.DELETE("/api/backends/:url", RemoveBackend)

	// run API server
	go func() {
		log.Printf("Starting API server on port 8082")
		apiRouter.Run(":8082")
	}()

	// run main server
	if useSSL {
		log.Printf("Starting HTTPS server on port %d\n", port)
		log.Fatal(server.ListenAndServeTLS(certPath, keyPath))
	} else {
		log.Printf("Starting HTTP server on port %d\n", port)
		log.Fatal(server.ListenAndServe())
	}
}
