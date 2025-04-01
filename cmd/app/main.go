package main

import (
	"net/http"

	"github.com/yingtu35/ShortenMe/internal/api"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", api.Home)
	mux.HandleFunc("/shorten", api.Shorten)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	server.ListenAndServe()
}
