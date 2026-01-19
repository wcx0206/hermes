package main

import (
	"context"
	"log"

	"github.com/wcx0206/hermes/internal/backup"
	"github.com/wcx0206/hermes/internal/config"
	"github.com/wcx0206/hermes/internal/logging"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	// Initialize logging
	err = logging.Init(cfg.Logging)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer logging.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Start cron server
	svr := backup.NewCronServer(cfg)
	if err := svr.Start(ctx); err != nil {
		log.Fatalf("start scheduler: %v", err)
	}
	select {}
}
