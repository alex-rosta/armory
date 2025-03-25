package interfaces

import (
	"context"
)

// SearchStore defines the interface for storing and retrieving search data
type SearchStore interface {
	// RecordSearch records a search in the store
	RecordSearch(ctx context.Context, region, realm, name string) error

	// GetRecentSearches gets the most recent searches
	GetRecentSearches(ctx context.Context) ([]SearchEntry, error)

	// Close closes the connection to the store
	Close() error
}

// SearchEntry represents a search entry
type SearchEntry struct {
	Character string `json:"character"`
	Realm     string `json:"realm"`
	Region    string `json:"region"`
	Timestamp string `json:"timestamp"`
}
