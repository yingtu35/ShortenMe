package api

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/yingtu35/ShortenMe/internal/config"
	"github.com/yingtu35/ShortenMe/internal/store"
)

type Handler struct {
	store  store.Store
	config config.Config
}

func NewHandler(store store.Store, config config.Config) *Handler {
	return &Handler{store: store, config: config}
}

type ShortenedURL struct {
	OriginalURL string
	ShortURL    string
}

type URLClickCounts struct {
	ShortURL   string
	ClickCount int64
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("internal/templates/index.html"))

	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	url := r.PostFormValue("url")
	if url == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
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

	tmpl := template.Must(template.ParseFiles("internal/templates/shorten.html"))

	err = tmpl.Execute(w, shortenedURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		tmpl := template.Must(template.ParseFiles("internal/templates/not-found.html"))
		err = tmpl.Execute(w, nil)
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
		tmpl := template.Must(template.ParseFiles("internal/templates/not-found.html"))
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	tmpl := template.Must(template.ParseFiles("internal/templates/url-click-counts.html"))

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
