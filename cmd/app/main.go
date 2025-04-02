package main

import (
	"log"
	"net/http"

	"github.com/yingtu35/ShortenMe/internal/api"
	"github.com/yingtu35/ShortenMe/internal/store"
)

func main() {
	// Create Redis store
	redisStore, err := store.NewRedisStore("redis-10241.c285.us-west-2-2.ec2.redns.redis-cloud.com:10241")
	if err != nil {
		log.Fatalf("Failed to create Redis store: %v", err)
	}
	defer redisStore.Close()

	// Create handler with store
	handler := api.NewHandler(redisStore)

	// Set up routes
	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", handler.Shorten)
	mux.HandleFunc("/click-counts", handler.URLClickCounts)
	mux.HandleFunc("/{shortURL}", handler.Redirect)
	mux.HandleFunc("/", handler.Home)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Starting server on port 8080")
	server.ListenAndServe()
}
