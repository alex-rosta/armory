package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"wowarmory/internal/interfaces"
)

// validRegions is the allowlist of Blizzard API regions; user-supplied region
// values are used to build API hostnames, so they must be validated.
var validRegions = map[string]bool{
	"us": true,
	"eu": true,
	"kr": true,
	"tw": true,
}

// ValidRegion reports whether region is an allowed Blizzard API region
func ValidRegion(region string) bool {
	return validRegions[region]
}

// BlizzardClient is a client for the Blizzard API
type BlizzardClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// Ensure BlizzardClient implements BlizzardAPI interface
var _ interfaces.BlizzardAPI = (*BlizzardClient)(nil)

// GetClientName returns the name of the client
func (c *BlizzardClient) GetClientName() string {
	return "BlizzardAPI"
}

// NewBlizzardClient creates a new Blizzard API client
func NewBlizzardClient(clientID, clientSecret string) *BlizzardClient {
	return &BlizzardClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAccessToken gets an access token from the Blizzard API
func (c *BlizzardClient) GetAccessToken() (string, error) {
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

// GetCharacterProfile gets a character profile from the Blizzard API
func (c *BlizzardClient) GetCharacterProfile(accessToken, region, realm, character string) (map[string]interface{}, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("missing access token")
	}
	if region == "" || realm == "" || character == "" {
		return nil, fmt.Errorf("missing region, realm, or character")
	}
	if !ValidRegion(region) {
		return nil, fmt.Errorf("invalid region: %q", region)
	}

	base := fmt.Sprintf("https://%s.api.blizzard.com/profile/wow/character/%s/%s",
		region, url.PathEscape(realm), url.PathEscape(character))
	query := fmt.Sprintf("?namespace=profile-%s&locale=en_US", region)

	endpoints := []string{
		base + query,
		base + "/character-media" + query,
		base + "/statistics" + query,
	}

	// Create a channel to receive responses from goroutines
	type apiResponse struct {
		data map[string]interface{}
		err  error
	}
	ch := make(chan apiResponse, len(endpoints))

	var wg sync.WaitGroup
	wg.Add(len(endpoints))

	for _, url := range endpoints {
		go func(url string) {
			defer wg.Done()
			data, err := c.fetchAPI(url, accessToken)
			ch <- apiResponse{data, err}
		}(url)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	// Combine the data from all endpoints
	combinedData := make(map[string]interface{})
	var firstError error

	for response := range ch {
		if response.err != nil {
			if firstError == nil {
				firstError = response.err
			}
			continue
		}
		for k, v := range response.data {
			combinedData[k] = v
		}
	}

	if firstError != nil && len(combinedData) == 0 {
		return nil, firstError
	}

	return combinedData, nil
}

// fetchAPI fetches data from a Blizzard API endpoint
func (c *BlizzardClient) fetchAPI(url, accessToken string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch API data: %s (status code: %d)", string(body), resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result, nil
}
