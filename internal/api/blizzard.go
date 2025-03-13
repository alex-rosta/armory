package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// BlizzardClient is a client for the Blizzard API
type BlizzardClient struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

// NewBlizzardClient creates a new Blizzard API client
func NewBlizzardClient(clientID, clientSecret string) *BlizzardClient {
	return &BlizzardClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{},
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

	// Define the endpoints to fetch data from
	endpoints := []string{
		fmt.Sprintf("https://%s.api.blizzard.com/profile/wow/character/%s/%s?namespace=profile-%s&locale=en_US&access_token=%s", region, realm, character, region, accessToken),
		fmt.Sprintf("https://%s.api.blizzard.com/profile/wow/character/%s/%s/character-media?namespace=profile-%s&locale=en_US&access_token=%s", region, realm, character, region, accessToken),
		fmt.Sprintf("https://%s.api.blizzard.com/profile/wow/character/%s/%s/statistics?namespace=profile-%s&locale=en_US&access_token=%s", region, realm, character, region, accessToken),
	}

	// Create a channel to receive responses from goroutines
	type apiResponse struct {
		data map[string]interface{}
		err  error
	}
	ch := make(chan apiResponse, len(endpoints))

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(len(endpoints))

	// Fetch data from each endpoint concurrently
	for _, url := range endpoints {
		go func(url string) {
			defer wg.Done()
			data, err := c.fetchAPI(url, accessToken)
			ch <- apiResponse{data, err}
		}(url)
	}

	// Wait for all goroutines to finish and close the channel
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
