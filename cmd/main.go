package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/nikhiljohn10/uagplugin/logger"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("%v", err)
		os.Exit(1)
	}
}

func init() {
	_ = godotenv.Load()
	if os.Getenv("UAG_ENV") == "development" {
		logger.SetDebugMode(true)
	}
}
