package api

import (
	"net/http"
)

// StaticHandler handles serving static pages
type StaticHandler struct {
	templateDir string
}

// NewStaticHandler creates a new StaticHandler
func NewStaticHandler(templateDir string) *StaticHandler {
	return &StaticHandler{
		templateDir: templateDir,
	}
}

// ServeTerms serves the Terms of Service page
func (h *StaticHandler) ServeTerms(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.templateDir+"/terms.html")
}

// ServePrivacy serves the Privacy Policy page
func (h *StaticHandler) ServePrivacy(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, h.templateDir+"/privacy.html")
}
