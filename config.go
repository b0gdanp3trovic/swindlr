package main

import (
	"log"

	"github.com/spf13/viper"
)

func initConfig(customPath string) {
	viper.SetConfigFile("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/swindlr/")
	viper.AddConfigPath("$HOME/.swindlr")
	viper.AddConfigPath(".")
	if customPath != "" {
		viper.AddConfigPath(customPath) // custom path from command-line argument
	}
	viper.AutomaticEnv()

	viper.SetDefault("port", 8080)
	viper.SetDefault("backends", []string{})

	//Read the config file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
}
