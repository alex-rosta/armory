package api

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"wowarmory/internal/interfaces"
)

// validRegions is the allowlist of Blizzard API regions; user-supplied region
// values are used to build API hostnames, so they must be validated.
var validRegions = map[string]bool{
	"us": true,
	"eu": true,
}

// ValidRegion reports whether region is an allowed Blizzard API region
func ValidRegion(region string) bool {
	return validRegions[region]
}

// BlizzardClient is a client for the Blizzard API
type BlizzardClient struct {
	tokens     *tokenSource
	httpClient *http.Client
	// apiURLFormat is formatted with the region to build the API base URL
	apiURLFormat string
}

// Ensure BlizzardClient implements BlizzardAPI interface
var _ interfaces.BlizzardAPI = (*BlizzardClient)(nil)

// NewBlizzardClient creates a new Blizzard API client
func NewBlizzardClient(clientID, clientSecret string) *BlizzardClient {
	return &BlizzardClient{
		tokens:       newTokenSource(clientID, clientSecret),
		httpClient:   newHTTPClient(),
		apiURLFormat: "https://%s.api.blizzard.com",
	}
}

// GetAccessToken gets an access token from the Blizzard API; tokens are cached until shortly before expiry
func (c *BlizzardClient) GetAccessToken() (string, error) {
	return c.tokens.Token()
}

// GetCharacterProfile gets a character profile from the Blizzard API.
// The profile endpoint is required; media and statistics are best-effort.
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

	base := fmt.Sprintf(c.apiURLFormat+"/profile/wow/character/%s/%s",
		region, url.PathEscape(realm), url.PathEscape(character))
	query := fmt.Sprintf("?namespace=profile-%s&locale=en_US", region)

	// Index 0 is the required profile; the rest only add optional fields
	endpoints := []string{
		base + query,
		base + "/character-media" + query,
		base + "/statistics" + query,
	}

	results := make([]map[string]interface{}, len(endpoints))
	errs := make([]error, len(endpoints))

	var wg sync.WaitGroup
	for i, endpoint := range endpoints {
		wg.Add(1)
		go func(i int, endpoint string) {
			defer wg.Done()
			results[i], errs[i] = c.fetchAPI(endpoint, accessToken)
		}(i, endpoint)
	}
	wg.Wait()

	if errs[0] != nil {
		return nil, errs[0]
	}

	combined := make(map[string]interface{})
	for i := len(results) - 1; i >= 0; i-- {
		if errs[i] != nil {
			log.Printf("optional character endpoint failed: %v", errs[i])
			continue
		}
		for k, v := range results[i] {
			combined[k] = v
		}
	}

	return combined, nil
}

// fetchAPI fetches data from a Blizzard API endpoint
func (c *BlizzardClient) fetchAPI(endpoint, accessToken string) (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	var result map[string]interface{}
	if err := getJSON(c.httpClient, req, &result); err != nil {
		return nil, err
	}
	return result, nil
}
