package interfaces

import (
	"context"
)

// APIClient defines the common interface for all API clients
type APIClient interface {
	// GetClientName returns the name of the client for logging/identification
	GetClientName() string
}

// BlizzardAPI defines the interface for Blizzard API operations
type BlizzardAPI interface {
	APIClient
	GetAccessToken() (string, error)
	GetCharacterProfile(accessToken, region, realm, character string) (map[string]interface{}, error)
}

// WarcraftLogsAPI defines the interface for WarcraftLogs API operations
type WarcraftLogsAPI interface {
	APIClient
	GetGuild(ctx context.Context, name, serverSlug, serverRegion string) (interface{}, error)
}
