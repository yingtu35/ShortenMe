package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// mockTimeProvider implements a time provider for testing
type mockTimeProvider struct {
	now time.Time
}

func (m *mockTimeProvider) Now() time.Time {
	return m.now
}

// setupTestRedis creates a new test Redis instance
func setupTestRedis(t *testing.T) *RedisStore {
	// Use a different database number for testing
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379" // default for local testing
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       1, // Use DB 1 for testing
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to connect to test Redis: %v", err)
	}

	// Clear the test database
	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("Failed to flush test database: %v", err)
	}

	// Create store with test client
	store := &RedisStore{
		client:       client,
		timeProvider: &mockTimeProvider{now: time.Now()},
	}

	// Register cleanup function
	t.Cleanup(func() {
		if err := client.FlushDB(ctx).Err(); err != nil {
			t.Errorf("Failed to flush test database: %v", err)
		}
		if err := client.Close(); err != nil {
			t.Errorf("Failed to close test Redis client: %v", err)
		}
	})

	return store
}

func TestCreateShortURL(t *testing.T) {
	store := setupTestRedis(t)

	tests := []struct {
		name        string
		originalURL string
		wantErr     bool
	}{
		{
			name:        "valid URL",
			originalURL: "https://example.com",
			wantErr:     false,
		},
		{
			name:        "empty URL",
			originalURL: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.CreateShortURL(tt.originalURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateShortURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == "" {
				t.Error("CreateShortURL() returned empty string for valid URL")
			}
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	store := setupTestRedis(t)

	// Create a test URL
	originalURL := "https://example.com"
	shortURL, err := store.CreateShortURL(originalURL)
	if err != nil {
		t.Fatalf("Failed to create test URL: %v", err)
	}

	// Extract the short code from the full URL
	shortCode := shortURL[len(os.Getenv("SHORTENME_URL"))+1:]

	tests := []struct {
		name     string
		shortURL string
		want     string
		wantErr  bool
	}{
		{
			name:     "existing URL",
			shortURL: shortCode,
			want:     originalURL,
			wantErr:  false,
		},
		{
			name:     "non-existent URL",
			shortURL: "nonexistent",
			want:     "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetOriginalURL(tt.shortURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOriginalURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetOriginalURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetClickCount(t *testing.T) {
	store := setupTestRedis(t)

	// Create a test URL
	originalURL := "https://example.com"
	shortURL, err := store.CreateShortURL(originalURL)
	if err != nil {
		t.Fatalf("Failed to create test URL: %v", err)
	}

	// Extract the short code from the full URL
	shortCode := shortURL[len(os.Getenv("SHORTENME_URL"))+1:]

	// Access the URL to increment click count
	_, err = store.GetOriginalURL(shortCode)
	if err != nil {
		t.Fatalf("Failed to access test URL: %v", err)
	}

	tests := []struct {
		name     string
		shortURL string
		want     int64
		wantErr  bool
	}{
		{
			name:     "existing URL with clicks",
			shortURL: shortCode,
			want:     1, // One click from the access above
			wantErr:  false,
		},
		{
			name:     "non-existent URL",
			shortURL: "nonexistent",
			want:     -1,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetClickCount(tt.shortURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetClickCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetClickCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPing(t *testing.T) {
	store := setupTestRedis(t)

	if err := store.Ping(); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}
