package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wcx0206/hermes/internal/backup"
	"github.com/wcx0206/hermes/internal/config"
	"github.com/wcx0206/hermes/internal/logging"
	"go.uber.org/zap"
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

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		first := true
		for {
			select {
			case sig := <-sigCh:
				if first {
					first = false
					logging.L().Info("shutdown signal received", zap.String("signal", sig.String()))
					cancel()
				} else {
					logging.L().Warn("second signal received, forcing exit", zap.String("signal", sig.String()))
					os.Exit(1)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Start cron server
	svr := backup.NewCronServer(cfg)
	if err := svr.Start(ctx); err != nil {
		log.Fatalf("failed start CronServer: %v", err)
		return
	}
	// Record pid
	if err := backup.SavePid(); err != nil {
		logging.L().Fatal("failed save pid", zap.Error(err))
		return
	}
	logging.L().Info("Hermes Backup Server started")

	<-ctx.Done()
	logging.L().Info("Hermes Backup Server stopped")
	if err := backup.RemovePid(); err != nil {
		logging.L().Error("failed remove pid", zap.Error(err))
	}
}
