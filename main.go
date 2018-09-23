package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mariaefi29/blog/config"
)

func main() {
	router := httprouter.New()
	router.GET("/", index)
	router.POST("/subscribe", subscribe)
	router.GET("/posts/show/:id", show)
	router.POST("/posts/show/:id", like)
	router.POST("/posts/show/:id/comments", comment)
	router.GET("/about", about)
	router.GET("/category/:category", category)
	router.GET("/contact", contact)
	router.POST("/contact", sendMessage)
	router.ServeFiles("/static/*filepath", http.Dir("public"))
	log.Fatal(http.ListenAndServe(":8080", router))
	defer config.Session.Close()
}
