package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"wowarmory/internal/interfaces"
)

type TokenClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

var _ interfaces.TokenAPI = (*TokenClient)(nil)

// GetClientName returns the name of the client
func (c *TokenClient) GetClientName() string {
	return "TokenAPI"
}

// NewTokenClient creates a new Token API client
func NewTokenClient(clientID, clientSecret string) *TokenClient {
	return &TokenClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{},
	}
}

// GetAccessToken gets an access token from the Token API
func (c *TokenClient) GetAccessToken() (string, error) {
	if c.clientID == "" || c.clientSecret == "" {
		return "", fmt.Errorf("missing client ID or client secret")
	}

	url := "https://oauth.battle.net/oauth/token"
	req, err := http.NewRequest("POST", url, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get access token: %s (status code: %d)", string(body), resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("unable to get access token from response")
	}
	return accessToken, nil
}

func (c *TokenClient) GetTokenPrice(accessToken, region string) (float64, error) {
	if accessToken == "" {
		return 0, fmt.Errorf("missing access token")
	}

	url := fmt.Sprintf("https://%s.api.blizzard.com/data/wow/token/index?namespace=dynamic-%s&locale=en_US", region, region)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to get token price: %s (status code: %d)", string(body), resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	price, ok := result["price"].(float64)
	if !ok {
		return 0, fmt.Errorf("unable to get token price from response")
	}
	return price, nil
}
