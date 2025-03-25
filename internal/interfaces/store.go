package interfaces

import (
	"context"
)

// SearchStore defines the interface for storing and retrieving search data
type SearchStore interface {
	// RecordSearch records a search in the store
	RecordSearch(ctx context.Context, searchType, region, realm, name string) error

	// GetRecentSearches gets the most recent searches
	GetRecentSearches(ctx context.Context) ([]SearchEntry, error)

	// Close closes the connection to the store
	Close() error
}

// SearchType represents the type of search
type SearchType string

const (
	// CharacterSearchType represents a character search
	CharacterSearchType SearchType = "character"

	// GuildSearchType represents a guild search
	GuildSearchType SearchType = "guild"
)

// SearchEntry represents a search entry
type SearchEntry struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Realm     string `json:"realm"`
	Region    string `json:"region"`
	Timestamp string `json:"timestamp"`
}
