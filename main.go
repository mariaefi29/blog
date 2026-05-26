package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/mariaefi29/blog/config"
	"github.com/mariaefi29/blog/internal/server"
	"github.com/mariaefi29/blog/store"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := store.New(ctx, cfg.Mongo.ConnectionString, cfg.Mongo.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := db.Disconnect(context.Background()); err != nil {
			log.Println(err)
		}
	}()

	srv := server.New(server.Params{
		Config: cfg,
		Store:  db,
	})

	stoppedCh := make(chan struct{})

	go func() {
		defer close(stoppedCh)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	log.Printf("http server address: http://localhost:%d", cfg.HTTP.Port)

	<-ctx.Done()

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}

	<-stoppedCh

	log.Print("http server stopped")
}
