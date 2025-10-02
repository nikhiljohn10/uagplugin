package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/nikhiljohn10/uagplugin/cmd"
	"github.com/nikhiljohn10/uagplugin/logger"
)

func main() {
	// Create a root context that cancels on SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := cmd.Root.ExecuteContext(ctx); err != nil {
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
