package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"
	"wowarmory/internal/config"
	"wowarmory/internal/interfaces"

	"github.com/redis/go-redis/v9"
)

const (
	// RecentSearchesKey is the sorted set holding recent search keys
	RecentSearchesKey = "recent_searches"

	MaxRecentSearches = 50

	SearchExpirationHours = 24
)

// Client is a wrapper around the Redis client that implements the SearchStore interface
type Client struct {
	rdb *redis.Client
}

// Ensure Client implements SearchStore interface
var _ interfaces.SearchStore = (*Client)(nil)

// SearchEntry represents a search entry
type SearchEntry struct {
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Realm     string    `json:"realm"`
	Region    string    `json:"region"`
	Timestamp time.Time `json:"timestamp"`
}

// NewClient creates a new Redis client
func NewClient(cfg *config.RedisConfig) (*Client, error) {
	var tlsConfig *tls.Config
	if cfg.UseTLS {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:      cfg.Addr,
		Password:  cfg.Password,
		DB:        cfg.DB,
		TLSConfig: tlsConfig,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Close closes the Redis client
func (c *Client) Close() error {
	return c.rdb.Close()
}

// RecordSearch records a search in Redis
func (c *Client) RecordSearch(ctx context.Context, searchType, region, realm, name string) error {
	now := time.Now()
	entry := SearchEntry{
		Type:      searchType,
		Name:      name,
		Realm:     realm,
		Region:    region,
		Timestamp: now,
	}

	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal search entry: %w", err)
	}

	key := fmt.Sprintf("%s:%s:%s:%s", searchType, region, realm, name)

	pipe := c.rdb.Pipeline()
	pipe.ZAdd(ctx, RecentSearchesKey, redis.Z{
		Score:  float64(now.Unix()),
		Member: key,
	})
	pipe.Set(ctx, key, entryJSON, time.Hour*SearchExpirationHours)
	// Keep only the most recent searches
	pipe.ZRemRangeByRank(ctx, RecentSearchesKey, 0, -MaxRecentSearches-1)
	pipe.Expire(ctx, RecentSearchesKey, time.Hour*SearchExpirationHours)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to record search: %w", err)
	}

	return nil
}

// GetRecentSearches gets the most recent searches, newest first
func (c *Client) GetRecentSearches(ctx context.Context) ([]interfaces.SearchEntry, error) {
	keys, err := c.rdb.ZRevRange(ctx, RecentSearchesKey, 0, MaxRecentSearches-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get recent searches: %w", err)
	}

	if len(keys) == 0 {
		return []interfaces.SearchEntry{}, nil
	}

	pipe := c.rdb.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	// Exec returns redis.Nil if any key has expired; that is handled per-command below
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get search entries: %w", err)
	}

	entries := make([]interfaces.SearchEntry, 0, len(keys))
	var stale []interface{}

	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			// Entry expired; remove its stale sorted set member
			stale = append(stale, keys[i])
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get search entry: %w", err)
		}

		var entry SearchEntry
		if err := json.Unmarshal([]byte(val), &entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshal search entry: %w", err)
		}

		entries = append(entries, interfaces.SearchEntry{
			Type:      entry.Type,
			Name:      entry.Name,
			Realm:     entry.Realm,
			Region:    entry.Region,
			Timestamp: entry.Timestamp.Format(time.RFC822),
		})
	}

	if len(stale) > 0 {
		// Best-effort cleanup; entries are still valid if this fails
		_ = c.rdb.ZRem(ctx, RecentSearchesKey, stale...).Err()
	}

	return entries, nil
}
