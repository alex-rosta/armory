package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"wowarmory/internal/interfaces"
	"wowarmory/internal/models"
)

// CharacterHandler handles character-related HTTP requests
type CharacterHandler struct {
	*BaseHandler
	blizzardClient interfaces.BlizzardAPI
}

// Ensure CharacterHandler implements Handler interface
var _ interfaces.Handler = (*CharacterHandler)(nil)

// NewCharacterHandler creates a new CharacterHandler
func NewCharacterHandler(base *BaseHandler, blizzardClient interfaces.BlizzardAPI) *CharacterHandler {
	return &CharacterHandler{
		BaseHandler:    base,
		blizzardClient: blizzardClient,
	}
}

// RegisterRoutes registers the handler's routes with the router
func (h *CharacterHandler) RegisterRoutes(router interfaces.RouteRegistrar) {
	router.HandleFunc("/", h.LookupCharacter)
}

// LookupCharacter shows the character search form and, when parameters are given, the result
func (h *CharacterHandler) LookupCharacter(w http.ResponseWriter, r *http.Request) {
	// "/" is a subtree pattern, so reject anything that is not the root path
	if r.URL.Path != "/" {
		h.RenderErrorPage(w, http.StatusNotFound, "Page not found", "The page you are looking for does not exist.")
		return
	}

	h.serveLookup(w, r, lookupConfig{
		Tab:        "character",
		Template:   "character_page",
		ResultKey:  "Character",
		Title:      "Character lookup",
		SearchType: interfaces.CharacterSearchType,
		NameParam:  "character",
		Form: SearchForm{
			Action:          "/",
			NameParam:       "character",
			NameLabel:       "Character name",
			NamePlaceholder: "e.g. tempests",
			SubmitLabel:     "Look up character",
		},
		ExternalURL: func(q Lookup) string { return blizzardArmoryURL(q.Region, q.Realm, q.Name) },
		Fetch: func(_ context.Context, q Lookup) (interface{}, string, error) {
			data, err := h.fetchCharacter(q)
			return data, data.Name, err
		},
	})
}

// fetchCharacter loads and processes a character from the Blizzard API
func (h *CharacterHandler) fetchCharacter(q Lookup) (models.CharacterData, error) {
	accessToken, err := h.blizzardClient.GetAccessToken()
	if err != nil {
		return models.CharacterData{}, fmt.Errorf("getting access token: %w", err)
	}

	profileData, err := h.blizzardClient.GetCharacterProfile(accessToken, q.Region, q.Realm, q.Name)
	if err != nil {
		return models.CharacterData{}, fmt.Errorf("getting character profile: %w", err)
	}

	return models.NewCharacterData(profileData, q.Region)
}

func blizzardArmoryURL(region, realm, character string) string {
	return fmt.Sprintf("https://worldofwarcraft.blizzard.com/en-gb/character/%s/%s/%s",
		url.PathEscape(region), url.PathEscape(realm), url.PathEscape(character))
}
