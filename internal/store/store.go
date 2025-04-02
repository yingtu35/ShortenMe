// Define the interface for the store that will be used by the API
package store

type Store interface {
	CreateShortURL(originalURL string) (string, error)
	GetOriginalURL(shortURL string) (string, error)
	GetClickCount(shortURL string) (int64, error)
}
