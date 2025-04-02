package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type URLData struct {
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	ClickCount  int64     `json:"click_count"`
}

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore() (*RedisStore, error) {
	addr := os.Getenv("REDIS_ADDR")
	username := os.Getenv("REDIS_USERNAME")
	password := os.Getenv("REDIS_PASSWORD")

	if addr == "" || username == "" || password == "" {
		return nil, fmt.Errorf("missing required Redis environment variables")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       0,
	})

	// Test the connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStore{client: client}, nil
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}

func (s *RedisStore) CreateShortURL(originalURL string) (string, error) {
	ctx := context.Background()

	// Get the next ID using Redis INCR
	id, err := s.client.Incr(ctx, "url_counter").Result()
	if err != nil {
		return "", fmt.Errorf("failed to increment counter: %w", err)
	}

	// Convert ID to base62 string
	shortURL := EncodeBase62(id)

	// Create URL data
	urlData := URLData{
		OriginalURL: originalURL,
		CreatedAt:   time.Now(),
		ClickCount:  0,
	}

	// Convert to JSON
	data, err := json.Marshal(urlData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal URL data: %w", err)
	}

	// Store the mapping in Redis
	err = s.client.Set(ctx, shortURL, data, 0).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store URL: %w", err)
	}

	fullShortURL := os.Getenv("SHORTENME_URL") + "/" + shortURL

	return fullShortURL, nil
}

func (s *RedisStore) GetOriginalURL(shortURL string) (string, error) {
	ctx := context.Background()

	// Get the URL data from Redis
	data, err := s.client.Get(ctx, shortURL).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	// Parse the URL data
	var urlData URLData
	if err := json.Unmarshal([]byte(data), &urlData); err != nil {
		return "", fmt.Errorf("failed to unmarshal URL data: %w", err)
	}

	// Increment click count
	urlData.ClickCount++
	updatedData, err := json.Marshal(urlData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal updated URL data: %w", err)
	}

	// Update Redis with new click count
	if err := s.client.Set(ctx, shortURL, updatedData, 0).Err(); err != nil {
		return "", fmt.Errorf("failed to update click count: %w", err)
	}

	return urlData.OriginalURL, nil
}

func (s *RedisStore) GetClickCount(shortURL string) (int64, error) {
	ctx := context.Background()

	// Get the URL data from Redis
	data, err := s.client.Get(ctx, shortURL).Result()
	if err == redis.Nil {
		return -1, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get URL: %w", err)
	}

	// Parse the URL data
	var urlData URLData
	if err := json.Unmarshal([]byte(data), &urlData); err != nil {
		return 0, fmt.Errorf("failed to unmarshal URL data: %w", err)
	}

	return urlData.ClickCount, nil
}
