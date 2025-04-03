package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/yingtu35/ShortenMe/internal/api"
	"github.com/yingtu35/ShortenMe/internal/config"
	"github.com/yingtu35/ShortenMe/internal/store"
)

func main() {
	// Only load .env file in development environment
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
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

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check Redis connection
		if err := redisStore.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Redis connection failed"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Serve favicon.ico directly
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/favicon.ico")
	})

	// Serve static files
	fs := http.FileServer(http.Dir("templates"))
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

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Channel to listen for errors coming from the server
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		log.Printf("Starting server on port %s", config.Port)
		serverErrors <- server.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking select waiting for either a server error or a signal
	select {
	case err := <-serverErrors:
		log.Printf("Error starting server: %v", err)

	case sig := <-shutdown:
		log.Printf("Start shutdown... Signal: %v", sig)

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Could not stop server gracefully: %v", err)
			if err := server.Close(); err != nil {
				log.Printf("Could not stop server: %v", err)
			}
		}
	}
}
