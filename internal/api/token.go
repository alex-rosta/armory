package api

import (
	"fmt"
	"net/http"
	"wowarmory/internal/interfaces"
)

// TokenClient is a client for the Blizzard WoW Token API
type TokenClient struct {
	tokens     *tokenSource
	httpClient *http.Client
	// apiURLFormat is formatted with the region to build the API base URL
	apiURLFormat string
}

var _ interfaces.TokenAPI = (*TokenClient)(nil)

// NewTokenClient creates a new Token API client
func NewTokenClient(clientID, clientSecret string) *TokenClient {
	return &TokenClient{
		tokens:       newTokenSource(clientID, clientSecret),
		httpClient:   newHTTPClient(),
		apiURLFormat: "https://%s.api.blizzard.com",
	}
}

// GetAccessToken gets an access token from the Blizzard API; tokens are cached until shortly before expiry
func (c *TokenClient) GetAccessToken() (string, error) {
	return c.tokens.Token()
}

// GetTokenPrice returns the WoW Token price in copper for a region
func (c *TokenClient) GetTokenPrice(accessToken, region string) (float64, error) {
	if accessToken == "" {
		return 0, fmt.Errorf("missing access token")
	}
	if !ValidRegion(region) {
		return 0, fmt.Errorf("invalid region: %q", region)
	}

	endpoint := fmt.Sprintf(c.apiURLFormat+"/data/wow/token/index?namespace=dynamic-%s&locale=en_US", region, region)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	var result struct {
		Price *float64 `json:"price"`
	}
	if err := getJSON(c.httpClient, req, &result); err != nil {
		return 0, fmt.Errorf("failed to get token price: %w", err)
	}
	if result.Price == nil {
		return 0, fmt.Errorf("unable to get token price from response")
	}
	return *result.Price, nil
}
