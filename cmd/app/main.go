package main

import (
	"net/http"

	"github.com/yingtu35/ShortenMe/internal/api"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", api.Shorten)
	mux.HandleFunc("/{shortURL}", api.Redirect)
	mux.HandleFunc("/", api.Home)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}
