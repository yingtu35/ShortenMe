package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/yingtu35/ShortenMe/internal/api"
	"github.com/yingtu35/ShortenMe/internal/config"
	"github.com/yingtu35/ShortenMe/internal/store"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	config := config.LoadConfig()

	// Create Redis store
	redisStore, err := store.NewRedisStore()
	if err != nil {
		log.Fatalf("Failed to create Redis store: %v", err)
	}
	defer redisStore.Close()

	// Create handler with store
	handler := api.NewHandler(redisStore, *config)

	// Set up routes
	mux := http.NewServeMux()

	// Serve static files
	fs := http.FileServer(http.Dir("internal/templates"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// API routes
	mux.HandleFunc("/shorten", handler.Shorten)
	mux.HandleFunc("/click-counts", handler.URLClickCounts)

	// Static pages
	mux.HandleFunc("/terms", api.ServeTerms)
	mux.HandleFunc("/privacy", api.ServePrivacy)

	// This should be the last route as it catches all other paths
	mux.HandleFunc("/{shortURL}", handler.Redirect)
	mux.HandleFunc("/", handler.Home)

	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: mux,
	}

	log.Println("Starting server on port " + config.Port)
	server.ListenAndServe()
}
