package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/akann77/wallet/internal/config"
	"github.com/akann77/wallet/internal/db"
	httpserver "github.com/akann77/wallet/internal/http"
	"github.com/akann77/wallet/internal/wallet"
)

func main() {
	cfgPath := os.Getenv("CONFIG_FILE")
	if cfgPath == "" {
		cfgPath = "config.env"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	database, err := db.Connect(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	if err := db.Migrate(ctx, database); err != nil {
		log.Fatal(err)
	}
	repo := wallet.NewRepository(database, cfg.TxMaxRetries)
	service := wallet.NewService(repo)
	addr := cfg.HTTPPort
	if !strings.HasPrefix(addr, ":") {
		addr = ":" + addr
	}
	server := &http.Server{
		Addr:         addr,
		Handler:      httpserver.NewRouter(service),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  90 * time.Second,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	waitForShutdown(server, cfg.ShutdownTimeout)
}

func waitForShutdown(server *http.Server, timeout time.Duration) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
