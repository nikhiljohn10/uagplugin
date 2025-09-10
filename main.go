package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/nikhiljohn10/uagplugin/cmd"
	"github.com/nikhiljohn10/uagplugin/logger"
)

func main() {
	if err := cmd.Root.Execute(); err != nil {
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
