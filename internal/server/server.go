package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Params struct {
	Port    int
	Timeout time.Duration

	StaticDir string
}

func New(params Params) *http.Server {
	router := chi.NewRouter()

	router.Get("/", index)
	router.Post("/subscribe", subscribe)
	router.Get("/posts/show/{id}", show)
	router.Post("/posts/show/{id}", like)
	router.Post("/posts/show/{id}/comments", comment)
	router.Get("/about", about)
	router.Get("/category/{category}", category)
	router.Get("/contact", contact)
	router.Post("/contact", sendMessage)

	staticDir := params.StaticDir
	if staticDir == "" {
		staticDir = "public"
	}
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	return &http.Server{
		Addr:              fmt.Sprintf(":%d", params.Port),
		Handler:           router,
		ReadHeaderTimeout: params.Timeout,
	}
}
