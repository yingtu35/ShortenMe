package api

import (
	"net/http"
)

// ServeTerms serves the Terms of Service page
func ServeTerms(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "internal/templates/terms.html")
}

// ServePrivacy serves the Privacy Policy page
func ServePrivacy(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "internal/templates/privacy.html")
}
