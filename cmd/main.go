package main

import (
	"avito_intern/internal/app"
	"avito_intern/internal/config"
	"avito_intern/internal/logger"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	cfg, err := config.MustLoad()
	if err != nil {
		log.Fatalf("Failed to load config: %s", err)
	}

	logger := logger.InitLogger(cfg)

	app := app.InitNewApp(context.Background(), cfg, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	app.Start(ctx)

	<-ctx.Done()
	ctxTimeOut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = app.Stop(ctxTimeOut)
	if err != nil {
		app.Logger.Error("Failed shutdown", "error", err)
		os.Exit(1)
	}

	app.Logger.Info("Gracefully stopped")
}
