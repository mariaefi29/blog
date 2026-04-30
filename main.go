package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mariaefi29/blog/config"
	"github.com/mariaefi29/blog/internal/server"
)

type httpConfig struct {
	Port    int
	Timeout time.Duration
}

func main() {
	defer config.Disconnect()

	cfg, err := loadHTTPConfig()
	if err != nil {
		log.Fatal(err)
	}
	srv := server.New(server.Params{
		Port:      cfg.Port,
		Timeout:   cfg.Timeout,
		StaticDir: "public",
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	stoppedCh := make(chan struct{})

	go func() {
		defer close(stoppedCh)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	log.Printf("http server address: http://localhost:%d", cfg.Port)

	<-ctx.Done()

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}

	<-stoppedCh

	log.Print("http server stopped")
}

func loadHTTPConfig() (httpConfig, error) {
	cfg := httpConfig{
		Port:    8080,
		Timeout: 30 * time.Second,
	}

	port := os.Getenv("PORT")
	if port == "" {
		return cfg, nil
	}

	parsedPort, err := strconv.Atoi(port)
	if err != nil {
		return httpConfig{}, fmt.Errorf("invalid PORT value %q: %w", port, err)
	}
	cfg.Port = parsedPort

	return cfg, nil
}
