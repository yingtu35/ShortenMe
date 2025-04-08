package api

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/yingtu35/ShortenMe/internal/config"
)

// mockStore implements the Store interface for testing
type mockStore struct {
	createShortURLFunc func(string) (string, error)
	getOriginalURLFunc func(string) (string, error)
	getClickCountFunc  func(string) (int64, error)
	pingFunc           func() error
}

func (m *mockStore) CreateShortURL(url string) (string, error) {
	if m.createShortURLFunc != nil {
		return m.createShortURLFunc(url)
	}
	return "", errors.New("CreateShortURL not implemented")
}

func (m *mockStore) GetOriginalURL(shortURL string) (string, error) {
	if m.getOriginalURLFunc != nil {
		return m.getOriginalURLFunc(shortURL)
	}
	return "", errors.New("GetOriginalURL not implemented")
}

func (m *mockStore) GetClickCount(shortURL string) (int64, error) {
	if m.getClickCountFunc != nil {
		return m.getClickCountFunc(shortURL)
	}
	return 0, errors.New("GetClickCount not implemented")
}

func (m *mockStore) Ping() error {
	if m.pingFunc != nil {
		return m.pingFunc()
	}
	return errors.New("Ping not implemented")
}

// getTemplateDir returns the absolute path to the templates directory
func getTemplateDir(t *testing.T) string {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Find the project root by looking for go.mod
	projectRoot := wd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Fatalf("Could not find project root directory")
		}
		projectRoot = parent
	}

	// Set the template directory path
	templateDir := filepath.Join(projectRoot, "templates")

	// Verify the template directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		t.Fatalf("Template directory does not exist: %v", templateDir)
	}

	return templateDir
}

func TestHome(t *testing.T) {
	// Get template directory
	templateDir := getTemplateDir(t)

	// Create a test config
	cfg := config.Config{
		BaseURL: "http://localhost:8080",
	}

	// Create a mock store
	mockStore := &mockStore{}

	// Create a handler with mock store and config
	handler := NewHandler(mockStore, cfg, templateDir)

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.Home(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body
	// Since we're testing a template, we can check for specific content
	// that should be present in the rendered template
	expected := []string{
		"<title>ShortenMe</title>",
		"<form",
		"url",
		"submit",
	}

	body := rr.Body.String()
	for _, content := range expected {
		if !strings.Contains(body, content) {
			t.Errorf("handler returned unexpected body: missing %v", content)
		}
	}
}

func TestShorten(t *testing.T) {
	// Get template directory
	templateDir := getTemplateDir(t)

	// Create a test config
	cfg := config.Config{
		BaseURL: "http://localhost:8080",
	}

	tests := []struct {
		name            string
		url             string
		mockShortURL    string
		mockError       error
		expectedStatus  int
		expectedContent []string
	}{
		{
			name:           "successful shortening",
			url:            "https://example.com",
			mockShortURL:   "http://localhost:8080/abc123",
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"Visit URL",
				"http://localhost:8080/abc123",
				"https://example.com",
			},
		},
		{
			name:           "empty URL",
			url:            "",
			mockShortURL:   "",
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedContent: []string{
				"URL is required",
			},
		},
		{
			name:           "store error",
			url:            "https://example.com",
			mockShortURL:   "",
			mockError:      errors.New("store error"),
			expectedStatus: http.StatusInternalServerError,
			expectedContent: []string{
				"store error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock store with the test case behavior
			mockStore := &mockStore{
				createShortURLFunc: func(url string) (string, error) {
					return tt.mockShortURL, tt.mockError
				},
			}

			// Create a handler with mock store and config
			handler := NewHandler(mockStore, cfg, templateDir)

			// Create form data
			form := bytes.NewBufferString("url=" + tt.url)

			// Create a test request
			req := httptest.NewRequest("POST", "/shorten", form)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler.Shorten(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// Check the response body
			body := rr.Body.String()
			for _, content := range tt.expectedContent {
				if !strings.Contains(body, content) {
					t.Errorf("handler returned unexpected body: missing %v", content)
				}
			}
		})
	}
}

func TestRedirect(t *testing.T) {
	// Get template directory
	templateDir := getTemplateDir(t)

	// Create a test config
	cfg := config.Config{
		BaseURL: "http://localhost:8080",
	}

	tests := []struct {
		name             string
		shortURL         string
		mockOriginalURL  string
		mockError        error
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:             "successful redirect",
			shortURL:         "abc123",
			mockOriginalURL:  "https://example.com",
			mockError:        nil,
			expectedStatus:   http.StatusFound,
			expectedLocation: "https://example.com",
		},
		{
			name:             "non-existent URL",
			shortURL:         "nonexistent",
			mockOriginalURL:  "",
			mockError:        nil,
			expectedStatus:   http.StatusOK,
			expectedLocation: "",
		},
		{
			name:             "store error",
			shortURL:         "abc123",
			mockOriginalURL:  "",
			mockError:        errors.New("store error"),
			expectedStatus:   http.StatusInternalServerError,
			expectedLocation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock store with the test case behavior
			mockStore := &mockStore{
				getOriginalURLFunc: func(shortURL string) (string, error) {
					return tt.mockOriginalURL, tt.mockError
				},
			}

			// Create a handler with mock store and config
			handler := NewHandler(mockStore, cfg, templateDir)

			// Create a chi router
			r := chi.NewRouter()

			// Set up the route with path parameter
			r.Get("/{shortURL}", handler.Redirect)

			// Create a test request
			req := httptest.NewRequest("GET", "/"+tt.shortURL, nil)

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Serve the request using the chi router
			r.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// Check the location header for redirects
			if tt.expectedLocation != "" {
				if location := rr.Header().Get("Location"); location != tt.expectedLocation {
					t.Errorf("handler returned wrong location: got %v want %v",
						location, tt.expectedLocation)
				}
			}

			// For non-existent URLs, check the not-found template
			if tt.mockOriginalURL == "" && tt.mockError == nil {
				expected := []string{
					"not found",
					tt.shortURL,
				}
				body := rr.Body.String()
				for _, content := range expected {
					if !strings.Contains(body, content) {
						t.Errorf("handler returned unexpected body: missing %v", content)
					}
				}
			}
		})
	}
}

func TestURLClickCounts(t *testing.T) {
	// Get template directory
	templateDir := getTemplateDir(t)

	// Create a test config
	cfg := config.Config{
		BaseURL: "http://localhost:8080",
	}

	tests := []struct {
		name            string
		shortURL        string
		mockClickCount  int64
		mockError       error
		expectedStatus  int
		expectedContent []string
	}{
		{
			name:           "successful click count",
			shortURL:       "http://localhost:8080/abc123",
			mockClickCount: 42,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"Click Count",
				"42",
				"abc123",
			},
		},
		{
			name:           "non-existent URL",
			shortURL:       "http://localhost:8080/nonexistent",
			mockClickCount: -1,
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedContent: []string{
				"not found",
				"nonexistent",
			},
		},
		{
			name:           "store error",
			shortURL:       "http://localhost:8080/abc123",
			mockClickCount: 0,
			mockError:      errors.New("store error"),
			expectedStatus: http.StatusInternalServerError,
			expectedContent: []string{
				"store error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock store with the test case behavior
			mockStore := &mockStore{
				getClickCountFunc: func(shortURL string) (int64, error) {
					return tt.mockClickCount, tt.mockError
				},
			}

			// Create a handler with mock store and config
			handler := NewHandler(mockStore, cfg, templateDir)

			// Create form data
			form := bytes.NewBufferString("shortURL=" + tt.shortURL)

			// Create a test request
			req := httptest.NewRequest("POST", "/click-count", form)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler.URLClickCounts(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// Check the response body
			body := rr.Body.String()
			for _, content := range tt.expectedContent {
				if !strings.Contains(body, content) {
					t.Errorf("handler returned unexpected body: missing %v", content)
				}
			}
		})
	}
}
