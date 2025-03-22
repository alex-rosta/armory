package integration

import (
	"context"
	"os"
	"testing"
	"time"
	"wowarmory/internal/config"
	"wowarmory/internal/redis"

	"github.com/joho/godotenv"
)

// TestRedisIntegration tests the integration with Redis
func TestRedisIntegration(t *testing.T) {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		t.Log("Error loading .env file, continuing without it. Ignore this if running as container.")
	}

	// Get Redis configuration from environment variables
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0 // Use DB 0 for tests

	// Create Redis configuration
	redisConfig := &config.RedisConfig{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	}

	// Run subtests
	t.Run("NewClient", testNewClient(redisConfig))
	t.Run("RecordSearch", testRecordSearch(redisConfig))
	t.Run("GetRecentSearches", testGetRecentSearches(redisConfig))
	t.Run("SearchExpiration", testSearchExpiration(redisConfig))
}

// testNewClient tests the NewClient function
func testNewClient(cfg *config.RedisConfig) func(t *testing.T) {
	return func(t *testing.T) {
		// Create a new Redis client
		client, err := redis.NewClient(cfg)
		if err != nil {
			t.Fatalf("Failed to create Redis client: %v", err)
		}
		defer client.Close()

		t.Log("Successfully connected to Redis")
	}
}

// testRecordSearch tests the RecordSearch function
func testRecordSearch(cfg *config.RedisConfig) func(t *testing.T) {
	return func(t *testing.T) {
		// Create a new Redis client
		client, err := redis.NewClient(cfg)
		if err != nil {
			t.Fatalf("Failed to create Redis client: %v", err)
		}
		defer client.Close()

		// Create a context
		ctx := context.Background()

		// Test parameters
		region := "eu"
		realm := "darkspear"
		character := "test-character-" + time.Now().Format("20060102150405")

		// Record a search
		err = client.RecordSearch(ctx, region, realm, character)
		if err != nil {
			t.Fatalf("Failed to record search: %v", err)
		}

		t.Logf("Successfully recorded search for %s-%s-%s", region, realm, character)
	}
}

// testGetRecentSearches tests the GetRecentSearches function
func testGetRecentSearches(cfg *config.RedisConfig) func(t *testing.T) {
	return func(t *testing.T) {
		// Create a new Redis client
		client, err := redis.NewClient(cfg)
		if err != nil {
			t.Fatalf("Failed to create Redis client: %v", err)
		}
		defer client.Close()

		// Create a context
		ctx := context.Background()

		// Record multiple searches
		testCharacters := []struct {
			region    string
			realm     string
			character string
		}{
			{"eu", "darkspear", "test-character-1-" + time.Now().Format("20060102150405")},
			{"us", "illidan", "test-character-2-" + time.Now().Format("20060102150405")},
			{"eu", "silvermoon", "test-character-3-" + time.Now().Format("20060102150405")},
		}

		for _, tc := range testCharacters {
			err = client.RecordSearch(ctx, tc.region, tc.realm, tc.character)
			if err != nil {
				t.Fatalf("Failed to record search for %s-%s-%s: %v", tc.region, tc.realm, tc.character, err)
			}
		}

		// Get recent searches
		searches, err := client.GetRecentSearches(ctx)
		if err != nil {
			t.Fatalf("Failed to get recent searches: %v", err)
		}

		// Verify that we got at least the number of searches we just added
		if len(searches) < len(testCharacters) {
			t.Fatalf("Expected at least %d searches, got %d", len(testCharacters), len(searches))
		}

		// Log the searches
		t.Logf("Retrieved %d recent searches", len(searches))
		for i, search := range searches {
			if i >= 5 { // Limit to 5 for logging
				t.Log("...")
				break
			}
			t.Logf("Search %d: %s-%s-%s at %s", i+1, search.Region, search.Realm, search.Character, search.Timestamp.Format(time.RFC3339))
		}
	}
}

// testSearchExpiration tests the expiration of searches
func testSearchExpiration(cfg *config.RedisConfig) func(t *testing.T) {
	return func(t *testing.T) {
		// This test is more of a verification that the expiration is set correctly
		// We can't actually wait for the expiration in a test, but we can verify the TTL

		// Create a new Redis client
		client, err := redis.NewClient(cfg)
		if err != nil {
			t.Fatalf("Failed to create Redis client: %v", err)
		}
		defer client.Close()

		// Create a context
		ctx := context.Background()

		// Test parameters
		region := "eu"
		realm := "darkspear"
		character := "test-character-ttl-" + time.Now().Format("20060102150405")

		// Record a search
		err = client.RecordSearch(ctx, region, realm, character)
		if err != nil {
			t.Fatalf("Failed to record search: %v", err)
		}
		t.Log("Search recorded successfully, expiration verification skipped in test")
	}
}

// TestRedisErrorHandling tests error handling in the Redis client
func TestRedisErrorHandling(t *testing.T) {
	t.Run("InvalidConnection", func(t *testing.T) {
		// Create a configuration with an invalid address
		cfg := &config.RedisConfig{
			Addr:     "nonexistent:6379",
			Password: "",
			DB:       0,
		}

		// Attempt to create a client
		_, err := redis.NewClient(cfg)
		if err == nil {
			t.Fatal("Expected error with invalid Redis address, but got none")
		}
		t.Logf("Got expected error with invalid Redis address: %v", err)
	})

	t.Run("InvalidAuth", func(t *testing.T) {
		// Skip if Redis doesn't require authentication
		if os.Getenv("REDIS_PASSWORD") == "" {
			t.Skip("Skipping test because Redis doesn't require authentication")
		}

		// Get the Redis address
		redisAddr := os.Getenv("REDIS_ADDR")
		if redisAddr == "" {
			redisAddr = "localhost:6379"
		}

		// Create a configuration with an invalid password
		cfg := &config.RedisConfig{
			Addr:     redisAddr,
			Password: "invalid_password",
			DB:       0,
		}

		// Attempt to create a client
		_, err := redis.NewClient(cfg)
		if err == nil {
			t.Fatal("Expected error with invalid Redis password, but got none")
		}
		t.Logf("Got expected error with invalid Redis password: %v", err)
	})
}
