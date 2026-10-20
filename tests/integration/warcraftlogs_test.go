package integration

import (
	"context"
	"os"
	"testing"
	"wowarmory/internal/api"

	"github.com/joho/godotenv"
)

// warcraftlogsToken returns the API token or skips the test when it is not configured
func warcraftlogsToken(t *testing.T) string {
	t.Helper()
	_ = godotenv.Load()

	token := os.Getenv("WARCRAFTLOGS_API_TOKEN")
	if token == "" {
		t.Skip("WARCRAFTLOGS_API_TOKEN not set")
	}
	return token
}

// TestWarcraftlogsAPIIntegration tests the integration with the Warcraftlogs API
func TestWarcraftlogsAPIIntegration(t *testing.T) {
	client := api.NewWarcraftlogsClient(warcraftlogsToken(t))

	// Using a known guild
	guildData, err := client.GetGuild(context.Background(), "divine intervention", "darkspear", "eu")
	if err != nil {
		t.Fatalf("Failed to get guild data: %v", err)
	}

	response, ok := guildData.(*api.GuildResponse)
	if !ok {
		t.Fatalf("Failed to cast guild data to GuildResponse: %T", guildData)
	}
	if response.GuildData.Guild.Name == "" {
		t.Fatal("Guild name is empty")
	}

	t.Logf("Guild %s has %d members", response.GuildData.Guild.Name, response.GuildData.Guild.Members.Total)
}

// TestWarcraftlogsAPIErrorHandling tests lookups that should not yield a guild
func TestWarcraftlogsAPIErrorHandling(t *testing.T) {
	// The API may either return an error or a null guild; both mean "no guild"
	noGuild := []struct {
		name                string
		guild, server, regn string
	}{
		{"UnknownGuild", "nonexistentguild123456789", "darkspear", "eu"},
		{"UnknownServer", "divine intervention", "nonexistentserver123456789", "eu"},
		{"EmptyGuild", "", "darkspear", "eu"},
		{"EmptyServer", "divine intervention", "", "eu"},
		{"EmptyRegion", "divine intervention", "darkspear", ""},
	}
	for _, tc := range noGuild {
		t.Run(tc.name, func(t *testing.T) {
			client := api.NewWarcraftlogsClient(warcraftlogsToken(t))

			resp, err := client.GetGuild(context.Background(), tc.guild, tc.server, tc.regn)
			if err != nil {
				return
			}
			if r, ok := resp.(*api.GuildResponse); !ok || r.GuildData.Guild.Name != "" {
				t.Fatalf("Expected no guild, got %+v", resp)
			}
		})
	}
}
