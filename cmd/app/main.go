package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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
	defer func() {
		if err := redisStore.Close(); err != nil {
			log.Printf("Error closing Redis store: %v", err)
		}
	}()

	// Get the absolute path to the templates directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	templateDir := filepath.Join(wd, "templates")

	// Create handler with store and template directory
	handler := api.NewHandler(redisStore, *config, templateDir)

	// Create static handler
	staticHandler := api.NewStaticHandler(templateDir)

	// Create chi router
	r := chi.NewRouter()

	// Add CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"chrome-extension://*"},
		AllowedMethods: []string{"POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}))

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check Redis connection
		if err := redisStore.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			if _, err := w.Write([]byte("Redis connection failed")); err != nil {
				log.Printf("Error writing response: %v", err)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})

	// Serve favicon.ico directly
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(templateDir, "favicon.ico"))
	})

	// Serve static files
	fs := http.FileServer(http.Dir(templateDir))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	// API routes for Chrome extension
	r.Post("/api/shorten", handler.APIShorten)

	// API routes
	r.Post("/shorten", handler.Shorten)
	r.Post("/click-counts", handler.URLClickCounts)

	// Static pages
	r.Get("/terms", staticHandler.ServeTerms)
	r.Get("/privacy", staticHandler.ServePrivacy)

	// image icon
	r.Get("/shortenme-icon.png", staticHandler.ServeStaticIcon)

	// This should be the last route as it catches all other paths
	r.Get("/{shortURL}", handler.Redirect)
	r.Get("/", handler.Home)

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      r,
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
