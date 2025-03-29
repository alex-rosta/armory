package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"wowarmory/internal/api"

	"github.com/joho/godotenv"
)

// TestWarcraftlogsAPIIntegration tests the integration with the Warcraftlogs API
func TestWarcraftlogsAPIIntegration(t *testing.T) {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Error loading .env file, continuing without it. Ignore this if running as container.")
	}

	// Get Warcraftlogs API token from environment variables
	accessToken := os.Getenv("WARCRAFTLOGS_API_TOKEN")
	if accessToken == "" {
		t.Fatal("WARCRAFTLOGS_API_TOKEN environment variable must be set")
	}

	t.Run("GetGuild", testGetGuild(accessToken))
}

// testGetGuild tests the GetGuild function
func testGetGuild(accessToken string) func(t *testing.T) {
	return func(t *testing.T) {
		// Create a new Warcraftlogs API client
		client := api.NewWarcraftlogsClient(accessToken)

		// Test parameters - using a known guild
		serverRegion := "eu"
		serverSlug := "darkspear"
		guildName := "divine intervention"

		// Set up context
		ctx := context.Background()

		// Get guild data
		guildData, err := client.GetGuild(ctx, guildName, serverSlug, serverRegion)
		if err != nil {
			t.Fatalf("Failed to get guild data: %v", err)
		}

		// Verify that we got a response
		if guildData == nil {
			t.Fatal("Guild data is nil")
		}

		// Type assertion to access the data
		response, ok := guildData.(*api.GuildResponse)
		if !ok {
			t.Fatalf("Failed to cast guild data to GuildResponse: %T", guildData)
		}

		// Verify the guild name
		if response.GuildData.Guild.Name == "" {
			t.Fatal("Guild name is empty")
		}

		t.Logf("Successfully obtained guild data for: %s", response.GuildData.Guild.Name)

		// Check that members total is available
		if response.GuildData.Guild.Members.Total <= 0 {
			t.Log("Warning: Guild members total is zero or negative")
		} else {
			t.Logf("Guild has %d members", response.GuildData.Guild.Members.Total)
		}
	}
}

// TestWarcraftlogsAPIErrorHandling tests error handling in the Warcraftlogs API client
func TestWarcraftlogsAPIErrorHandling(t *testing.T) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Get valid access token
	accessToken := os.Getenv("WARCRAFTLOGS_API_TOKEN")

	t.Run("InvalidGuildName", func(t *testing.T) {
		if accessToken == "" {
			t.Skip("Skipping test due to missing access token")
		}

		// Create a client with valid token
		client := api.NewWarcraftlogsClient(accessToken)

		// Set up context
		ctx := context.Background()

		// Attempt to get a non-existent guild
		_, err := client.GetGuild(ctx, "nonexistentguild123456789", "darkspear", "eu")
		if err != nil {
			t.Fatal("Expected error with non-existent guild, but got none")
		}
		t.Logf("Got expected error with non-existent guild: %v", err)
	})

	t.Run("InvalidServerName", func(t *testing.T) {
		if accessToken == "" {
			t.Skip("Skipping test due to missing access token")
		}

		// Create a client with valid token
		client := api.NewWarcraftlogsClient(accessToken)

		// Set up context
		ctx := context.Background()

		// Attempt to get a guild on a non-existent server
		_, err := client.GetGuild(ctx, "divine intervention", "nonexistentserver123456789", "eu")
		if err != nil {
			t.Fatal("Expected error with non-existent server, but got none")
		}
		t.Logf("Got expected error with non-existent server: %v", err)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		// Create a client with invalid token
		client := api.NewWarcraftlogsClient("invalid_token")

		// Set up context
		ctx := context.Background()

		// Attempt to get guild data
		_, err := client.GetGuild(ctx, "divine intervention", "darkspear", "eu")
		if err != nil {
			t.Fatal("Expected error with invalid token, but got none")
		}
		t.Logf("Got expected error with invalid token: %v", err)
	})

	t.Run("EmptyParameters", func(t *testing.T) {
		if accessToken == "" {
			t.Skip("Skipping test due to missing access token")
		}

		// Create a client with valid token
		client := api.NewWarcraftlogsClient(accessToken)

		// Set up context
		ctx := context.Background()

		// Test with empty guild name
		_, err := client.GetGuild(ctx, "", "darkspear", "eu")
		if err != nil {
			t.Fatal("Expected error with empty guild name, but got none")
		}

		// Test with empty server slug
		_, err = client.GetGuild(ctx, "divine intervention", "", "eu")
		if err != nil {
			t.Fatal("Expected error with empty server slug, but got none")
		}

		// Test with empty region
		_, err = client.GetGuild(ctx, "divine intervention", "darkspear", "")
		if err != nil {
			t.Fatal("Expected error with empty region, but got none")
			t.Logf("Got expected error with empty region: %v", err)
		}
	})
}
