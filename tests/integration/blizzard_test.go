package integration

import (
	"fmt"
	"os"
	"testing"
	"wowarmory/internal/api"

	"github.com/joho/godotenv"
)

// TestBlizzardAPIIntegration tests the integration with the Blizzard API
func TestBlizzardAPIIntegration(t *testing.T) {

	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Error loading .env file, continuing without it. Ignore this if running as container.")
	}
	// Get Blizzard API credentials from environment variables
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		t.Fatal("CLIENT_ID and CLIENT_SECRET environment variables must be set")
	}

	t.Run("GetAccessToken", testGetAccessToken(clientID, clientSecret))
	t.Run("GetCharacterProfile", testGetCharacterProfile(clientID, clientSecret))
}

// testGetAccessToken tests the GetAccessToken function
func testGetAccessToken(clientID, clientSecret string) func(t *testing.T) {
	return func(t *testing.T) {
		// Create a new Blizzard API client
		client := api.NewBlizzardClient(clientID, clientSecret)

		// Get an access token
		token, err := client.GetAccessToken()
		if err != nil {
			t.Fatalf("Failed to get access token: %v", err)
		}

		// Verify that the token is not empty
		if token == "" {
			t.Fatal("Access token is empty")
		}

		t.Logf("Successfully obtained access token: %s...", token[:10])
	}
}

// testGetCharacterProfile tests the GetCharacterProfile function
func testGetCharacterProfile(clientID, clientSecret string) func(t *testing.T) {
	return func(t *testing.T) {
		// Create a new Blizzard API client
		client := api.NewBlizzardClient(clientID, clientSecret)

		// Get an access token
		token, err := client.GetAccessToken()
		if err != nil {
			t.Fatalf("Failed to get access token: %v", err)
		}

		// Test parameters - using a known character
		region := "eu"
		realm := "darkspear"
		character := "tempests"

		// Get character profile
		profile, err := client.GetCharacterProfile(token, region, realm, character)
		if err != nil {
			t.Fatalf("Failed to get character profile: %v", err)
		}

		// Verify that the profile contains expected fields
		requiredFields := []string{"name", "level", "average_item_level", "character_class"}
		for _, field := range requiredFields {
			if _, ok := profile[field]; !ok {
				t.Errorf("Character profile missing required field: %s", field)
			}
		}

		// Log some basic character information
		t.Logf("Character: %s", profile["name"])
		if level, ok := profile["level"].(float64); ok {
			t.Logf("Level: %.0f", level)
		}
		if itemLevel, ok := profile["average_item_level"].(float64); ok {
			t.Logf("Item Level: %.0f", itemLevel)
		}
	}
}

// TestGetTokenPrice tests the GetTokenPrice function
func TestGetTokenPrice(t *testing.T) {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Error loading .env file, continuing without it. Ignore this if running as container.")
	}
	// Get Blizzard API credentials from environment variables
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		t.Fatal("CLIENT_ID and CLIENT_SECRET environment variables must be set")
	}

	// Create a new Blizzard API client
	client := api.NewTokenClient(clientID, clientSecret)

	// Get an access token
	token, err := client.GetAccessToken()
	if err != nil {
		t.Fatalf("Failed to get access token: %v", err)
	}

	region := "eu"

	// Get token price
	price, err := client.GetTokenPrice(token, region)
	if err != nil {
		t.Fatalf("Failed to get token price: %v", err)
	}

	t.Logf("Token price in %s: %.0f", region, price)
}

// TestBlizzardAPIErrorHandling tests error handling in the Blizzard API client
func TestBlizzardAPIErrorHandling(t *testing.T) {
	t.Run("InvalidCredentials", func(t *testing.T) {
		// Create a client with invalid credentials
		client := api.NewBlizzardClient("invalid_id", "invalid_secret")

		// Attempt to get an access token
		_, err := client.GetAccessToken()
		if err == nil {
			t.Fatal("Expected error with invalid credentials, but got none")
		}
		t.Logf("Got expected error with invalid credentials: %v", err)
	})

	t.Run("InvalidCharacter", func(t *testing.T) {
		// Get valid credentials
		clientID := os.Getenv("CLIENT_ID")
		clientSecret := os.Getenv("CLIENT_SECRET")

		if clientID == "" || clientSecret == "" {
			t.Skip("Skipping test due to missing credentials")
		}

		// Create a client with valid credentials
		client := api.NewBlizzardClient(clientID, clientSecret)

		// Get a valid access token
		token, err := client.GetAccessToken()
		if err != nil {
			t.Fatalf("Failed to get access token: %v", err)
		}

		// Attempt to get a non-existent character
		_, err = client.GetCharacterProfile(token, "eu", "silvermoon", "nonexistentcharacter123456789")
		if err == nil {
			t.Fatal("Expected error with non-existent character, but got none")
		}
		t.Logf("Got expected error with non-existent character: %v", err)
	})

	t.Run("MissingParameters", func(t *testing.T) {
		// Create a client
		client := api.NewBlizzardClient("id", "secret")

		// Test with empty access token
		_, err := client.GetCharacterProfile("", "eu", "silvermoon", "thrall")
		if err == nil {
			t.Fatal("Expected error with empty access token, but got none")
		}

		// Test with empty region
		_, err = client.GetCharacterProfile("token", "", "silvermoon", "thrall")
		if err == nil {
			t.Fatal("Expected error with empty region, but got none")
		}

		// Test with empty realm
		_, err = client.GetCharacterProfile("token", "eu", "", "thrall")
		if err == nil {
			t.Fatal("Expected error with empty realm, but got none")
		}

		// Test with empty character
		_, err = client.GetCharacterProfile("token", "eu", "silvermoon", "")
		if err == nil {
			t.Fatal("Expected error with empty character, but got none")
		}
	})
}
