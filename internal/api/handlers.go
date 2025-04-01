package api

import (
	"html/template"
	"net/http"
)

type ShortenedURL struct {
	OriginalURL string
	ShortURL    string
}

func Home(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("internal/templates/index.html"))

	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	url := r.PostFormValue("url")
	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	shortURL := generateShortURL(url)

	ShortenedURL := ShortenedURL{
		OriginalURL: url,
		ShortURL:    shortURL,
	}

	tmpl := template.Must(template.ParseFiles("internal/templates/shorten.html"))

	err := tmpl.Execute(w, ShortenedURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Create a dummy function to generate short URLs
func generateShortURL(url string) string {
	return "http://localhost:8080/" + "123"
}
