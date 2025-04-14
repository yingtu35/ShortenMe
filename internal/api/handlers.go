package api

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/yingtu35/ShortenMe/internal/config"
	"github.com/yingtu35/ShortenMe/internal/store"
)

type Handler struct {
	store       store.Store
	config      config.Config
	templateDir string
}

func NewHandler(store store.Store, config config.Config, templateDir string) *Handler {
	return &Handler{
		store:       store,
		config:      config,
		templateDir: templateDir,
	}
}

type ShortenedURL struct {
	OriginalURL string
	ShortURL    string
}

type URLClickCounts struct {
	ShortURL   string
	ClickCount int64
}

type NotFound struct {
	ShortURL string
}

// IsValidURL checks if the given string is a valid URL
func IsValidURL(input string) bool {
	parsedURL, err := url.ParseRequestURI(input)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}
	return true
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(h.templateDir + "/index.html"))

	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	url := r.PostFormValue("url")
	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if !IsValidURL(url) {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	shortURL, err := h.store.CreateShortURL(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	shortenedURL := ShortenedURL{
		OriginalURL: url,
		ShortURL:    shortURL,
	}

	tmpl := template.Must(template.ParseFiles(h.templateDir + "/shorten.html"))

	err = tmpl.Execute(w, shortenedURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	select {
	case <-ctx.Done():
		return
	default:
	}
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortURL := r.PathValue("shortURL")
	if shortURL == "" {
		http.Error(w, "Short URL is required", http.StatusBadRequest)
		return
	}

	originalURL, err := h.store.GetOriginalURL(shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if originalURL == "" {
		tmpl := template.Must(template.ParseFiles(h.templateDir + "/not-found.html"))
		err = tmpl.Execute(w, NotFound{ShortURL: shortURL})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func (h *Handler) URLClickCounts(w http.ResponseWriter, r *http.Request) {
	fullShortURL := r.FormValue("shortURL")
	shortURL := strings.TrimPrefix(fullShortURL, h.config.BaseURL+"/")
	if shortURL == "" {
		http.Error(w, "Short URL is required", http.StatusBadRequest)
		return
	}

	clickCount, err := h.store.GetClickCount(shortURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if clickCount == -1 {
		tmpl := template.Must(template.ParseFiles(h.templateDir + "/not-found.html"))
		err = tmpl.Execute(w, NotFound{ShortURL: shortURL})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	tmpl := template.Must(template.ParseFiles(h.templateDir + "/url-click-counts.html"))

	urlClickCounts := URLClickCounts{
		ShortURL:   fullShortURL,
		ClickCount: clickCount,
	}
	err = tmpl.Execute(w, urlClickCounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) APIShorten(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		h.respondWithJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Method not allowed"})
		return
	}

	var requestBody struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	url := requestBody.URL
	if url == "" {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "URL is required"})
		return
	}

	if !IsValidURL(url) {
		h.respondWithJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid URL"})
		return
	}

	shortURL, err := h.store.CreateShortURL(url)
	if err != nil {
		h.respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response := map[string]string{
		"original_url": url,
		"short_url":    shortURL,
	}

	h.respondWithJSON(w, http.StatusOK, response)

	select {
	case <-ctx.Done():
		return
	default:
	}
}

func (h *Handler) respondWithJSON(w http.ResponseWriter, statusCode int, payload map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
