package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func GetAccessToken() (string, error) {
	clientID, clientSecret := os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("missing client ID or client secret")
	}

	url := "https://oauth.battle.net/oauth/token"
	req, err := http.NewRequest("POST", url, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("unable to get access token")
	}
	return accessToken, nil
}

func GetCharacterProfile(accessToken, region, realm, character string) (map[string]interface{}, error) {
	endpoints := []string{
		fmt.Sprintf("https://%s.api.blizzard.com/profile/wow/character/%s/%s?namespace=profile-%s&locale=en_US&access_token=%s", region, realm, character, region, accessToken),
		fmt.Sprintf("https://%s.api.blizzard.com/profile/wow/character/%s/%s/character-media?namespace=profile-%s&locale=en_US&access_token=%s", region, realm, character, region, accessToken),
		fmt.Sprintf("https://%s.api.blizzard.com/profile/wow/character/%s/%s/statistics?namespace=profile-%s&locale=en_US&access_token=%s", region, realm, character, region, accessToken),
	}

	type apiResponse struct {
		data map[string]interface{}
		err  error
	}

	ch := make(chan apiResponse, len(endpoints))

	for _, url := range endpoints {
		go func(url string) {
			data, err := FetchAPI(url, accessToken)
			ch <- apiResponse{data, err}
		}(url)
	}

	combinedData := make(map[string]interface{})
	for range endpoints {
		response := <-ch
		if response.err != nil {
			return nil, response.err
		}
		for k, v := range response.data {
			combinedData[k] = v
		}
	}
	return combinedData, nil
}

func FetchAPI(url, accessToken string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
