package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	blizzardTokenURL = "https://oauth.battle.net/oauth/token"

	// maxResponseBytes bounds upstream response bodies read into memory
	maxResponseBytes = 1 << 20

	// tokenExpiryMargin refreshes the token slightly before it actually expires
	tokenExpiryMargin = time.Minute

	// tokenFailureBackoff is how long a failed token request is reused before trying again
	tokenFailureBackoff = 5 * time.Second
)

// ErrNotFound is returned when an upstream API responds with 404
var ErrNotFound = errors.New("not found")

// newHTTPClient returns an HTTP client for upstream APIs
func newHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

// getJSON performs req and decodes a 200 response body into v
func getJSON(client *http.Client, req *http.Request, v interface{}) error {
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("unexpected response: %s (status code: %d)", body, resp.StatusCode)
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseBytes)).Decode(v); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	return nil
}

// tokenSource fetches and caches Blizzard client-credentials access tokens
type tokenSource struct {
	clientID     string
	clientSecret string
	tokenURL     string
	httpClient   *http.Client

	mu          sync.Mutex
	token       string
	expiry      time.Time
	lastErr     error
	failedUntil time.Time
}

func newTokenSource(clientID, clientSecret string) *tokenSource {
	return &tokenSource{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     blizzardTokenURL,
		httpClient:   newHTTPClient(),
	}
}

// Token returns a cached access token, requesting a new one when it is missing or about to expire
func (s *tokenSource) Token() (string, error) {
	if s.clientID == "" || s.clientSecret == "" {
		return "", fmt.Errorf("missing client ID or client secret")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.token != "" && time.Now().Before(s.expiry) {
		return s.token, nil
	}
	// After a failure, answer from the recorded error so queued requests don't each wait out a full timeout
	if s.lastErr != nil && time.Now().Before(s.failedUntil) {
		return "", s.lastErr
	}

	token, expiresIn, err := s.fetch()
	if err != nil {
		s.lastErr, s.failedUntil = err, time.Now().Add(tokenFailureBackoff)
		return "", err
	}
	s.lastErr = nil

	s.token = token
	s.expiry = time.Now().Add(expiresIn - tokenExpiryMargin)
	return s.token, nil
}

// fetch requests a new token; callers hold s.mu
func (s *tokenSource) fetch() (string, time.Duration, error) {
	req, err := http.NewRequest(http.MethodPost, s.tokenURL, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.SetBasicAuth(s.clientID, s.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := getJSON(s.httpClient, req, &result); err != nil {
		return "", 0, fmt.Errorf("failed to get access token: %w", err)
	}
	if result.AccessToken == "" {
		return "", 0, fmt.Errorf("unable to get access token from response")
	}
	return result.AccessToken, time.Duration(result.ExpiresIn) * time.Second, nil
}
