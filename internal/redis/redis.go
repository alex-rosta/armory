package redis

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"wowarmory/internal/config"

	"github.com/redis/go-redis/v9"
)

const (
	// RecentSearchesKey is the key for the sorted set of recent searches
	RecentSearchesKey = "recent_searches"

	// MaxRecentSearches is the maximum number of recent searches to keep
	MaxRecentSearches = 50

	// SearchExpirationHours is the number of hours to keep searches
	SearchExpirationHours = 24
)

// Client is a wrapper around the Redis client
type Client struct {
	rdb *redis.Client
}

// SearchEntry represents a character search entry
type SearchEntry struct {
	Character string    `json:"character"`
	Realm     string    `json:"realm"`
	Region    string    `json:"region"`
	Timestamp time.Time `json:"timestamp"`
}

// NewClient creates a new Redis client
func NewClient(cfg *config.RedisConfig) (*Client, error) {
	var tlsConfig *tls.Config

	if os.Getenv("REDIS_CLOUD") == "true" {
		tlsConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		// TLS for cloud redis
		TLSConfig: tlsConfig,
	})

	// Test the connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Close closes the Redis client
func (c *Client) Close() error {
	return c.rdb.Close()
}

// RecordSearch records a character search in Redis
func (c *Client) RecordSearch(ctx context.Context, region, realm, character string) error {
	// Create a search entry
	entry := SearchEntry{
		Character: character,
		Realm:     realm,
		Region:    region,
		Timestamp: time.Now(),
	}

	// Serialize the entry to JSON
	entryJSON, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal search entry: %w", err)
	}

	// Add the entry to the sorted set with the current timestamp as the score
	score := float64(time.Now().Unix())
	key := fmt.Sprintf("%s:%s:%s", region, realm, character)

	// Use a pipeline to execute multiple commands atomically
	pipe := c.rdb.Pipeline()

	// Add to sorted set
	pipe.ZAdd(ctx, RecentSearchesKey, redis.Z{
		Score:  score,
		Member: key,
	})

	// Store the JSON data
	pipe.Set(ctx, key, entryJSON, time.Hour*SearchExpirationHours)

	// Trim the sorted set to keep only the most recent searches
	pipe.ZRemRangeByRank(ctx, RecentSearchesKey, 0, -MaxRecentSearches-1)

	// Expire sorted set on the most recent search entry
	pipe.Expire(ctx, RecentSearchesKey, time.Hour*SearchExpirationHours)

	// Execute the pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record search: %w", err)
	}

	return nil
}

// GetRecentSearches gets the most recent character searches
func (c *Client) GetRecentSearches(ctx context.Context) ([]SearchEntry, error) {
	// Get the most recent searches from the sorted set (highest scores first)
	keys, err := c.rdb.ZRevRange(ctx, RecentSearchesKey, 0, MaxRecentSearches-1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get recent searches: %w", err)
	}

	// If there are no searches, return an empty slice
	if len(keys) == 0 {
		return []SearchEntry{}, nil
	}

	// Get the JSON data for each key
	pipe := c.rdb.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get search entries: %w", err)
	}

	// Parse the JSON data into search entries
	entries := make([]SearchEntry, 0, len(keys))

	for i, cmd := range cmds {
		// Skip entries that no longer exist (may have expired)
		val, err := cmd.Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get search entry: %w", err)
		}

		var entry SearchEntry
		if err := json.Unmarshal([]byte(val), &entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshal search entry: %w", err)
		}
		// Expire search entries after a specified time, delete them
		if time.Since(entry.Timestamp) > time.Hour*SearchExpirationHours {
			if err := c.rdb.Del(ctx, keys[i]).Err(); c.rdb.ZRem(ctx, RecentSearchesKey, keys[i]).Err() != nil {
				return nil, fmt.Errorf("failed to delete expired search entry: %w", err)
			}
			continue
		}

		entries = append(entries, entry)
	}

	return entries, nil

}
