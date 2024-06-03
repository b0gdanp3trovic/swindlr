package main

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

// TODO - find a good way to check for invalid options

func checkSSLConfig() {
	useSSL := viper.GetBool("use_ssl")
	sslCertFile := viper.GetString("ssl_cert_file")
	sslKeyFile := viper.GetString("ssl_key_file")

	if useSSL {
		log.Println("SSL is enabled.")

		if sslCertFile == "" || sslKeyFile == "" {
			log.Fatal("Error: SSL is enabled but 'ssl_cert_file' or 'ssl_key_file' file is not specified.")
		}

		// Check if the certificate and key files exist
		if _, err := os.Stat(sslCertFile); os.IsNotExist(err) {
			log.Fatalf("Error: SSL certificate file '%s' not found.", sslCertFile)
		}
		if _, err := os.Stat(sslKeyFile); os.IsNotExist(err) {
			log.Fatalf("Error: SSL key file '%s' not found.", sslKeyFile)
		}

		log.Println("SSL configuration validated successfully.")
	} else {
		log.Println("SSL is not enabled.")
	}
}

func checkLoadBalancerStrategy() {
	strategy := viper.GetString("load_balancer.strategy")
	validStrategies := map[string]bool{
		"round_robin":       true,
		"least_connections": true,
		"random":            true,
		"latency_aware":     true,
	}

	if _, valid := validStrategies[strategy]; !valid {
		log.Fatalf("Invalid load balancing strategy specified: %s", strategy)
	}

	log.Printf("Load balancing strategy '%s' is set.", strategy)
}

func initConfig(customPath string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AddConfigPath("/etc/swindlr/")
	viper.AddConfigPath("$HOME/.swindlr")
	viper.AddConfigPath(".")
	if customPath != "" {
		viper.AddConfigPath(customPath)
	}
	viper.AutomaticEnv()

	// CONFIG VALUES
	viper.SetDefault("port", 8080)
	viper.SetDefault("backends", []string{})
	viper.SetDefault("use_ssl", false)
	viper.SetDefault("ssl_cert_file", "")
	viper.SetDefault("ssl_key_file", "")
	viper.SetDefault("use_dynamic", false)
	viper.SetDefault("apiPort", 8082)
	viper.SetDefault("load_balancer.strategy", "round_robin")
	viper.SetDefault("use_sticky_sessions", false)
	viper.SetDefault("rate_limiting.rate", 10.0)
	viper.SetDefault("rate_limiting.bucket_size", 5)
	viper.SetDefault("use_geo_routing", false)
	viper.SetDefault("use_cache", false)

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %s", err)
	} else {
		log.Printf("Loaded configuration from file: %s", viper.ConfigFileUsed())
	}

	//Validate
	checkSSLConfig()
	checkLoadBalancerStrategy()
}
