package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/b0gdanp3trovic/swindlr/api"
	"github.com/b0gdanp3trovic/swindlr/loadbalancer"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	Attempts int = iota
	Retry
)

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
	useDynamic := viper.GetBool("use_dynamic")
	strategy := viper.GetString("load_balancer.strategy")

	serverPool := loadbalancer.SetupServerPool(backendURLs, strategy)

	cache := loadbalancer.NewCache(5 * time.Minute)

	server := http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loadbalancer.LB(w, r, serverPool, cache)
		}),
	}

	go loadbalancer.Health(serverPool)
	go loadbalancer.ManageHealthUpdate()

	// Prepare API endpoints
	if useDynamic {
		gin.SetMode(gin.ReleaseMode)
		apiRouter := gin.Default()
		apiRouter.POST("/api/backends", func(c *gin.Context) {
			api.AddBackend(c, serverPool)
		})
		apiRouter.DELETE("/api/backends/:url", func(c *gin.Context) {
			api.RemoveBackend(c, serverPool)
		})

		// run API server
		go func() {
			log.Printf("Starting API server on port 8082")
			apiRouter.Run(":8082")
		}()

		log.Printf("Dynamic server pool management is enabled.")
	} else {
		log.Printf("Dynamic server pool management is disabled.")
	}

	// run main server
	if useSSL {
		log.Printf("Starting HTTPS server on port %d\n", port)
		log.Fatal(server.ListenAndServeTLS(certPath, keyPath))
	} else {
		log.Printf("Starting HTTP server on port %d\n", port)
		log.Fatal(server.ListenAndServe())
	}
}
