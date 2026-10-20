package interfaces

import (
	"context"
)

// BlizzardAPI defines the interface for Blizzard API operations
type BlizzardAPI interface {
	GetAccessToken() (string, error)
	GetCharacterProfile(accessToken, region, realm, character string) (map[string]interface{}, error)
}

// WarcraftLogsAPI defines the interface for WarcraftLogs API operations
type WarcraftLogsAPI interface {
	GetGuild(ctx context.Context, name, serverSlug, serverRegion string) (interface{}, error)
}

// TokenAPI defines the interface for Token API operations
type TokenAPI interface {
	GetAccessToken() (string, error)
	GetTokenPrice(accessToken, region string) (float64, error)
}
