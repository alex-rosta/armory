package handlers

import (
	"fmt"
	"log"
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
	router.HandleFunc("/character", h.GetCharacterTemplate)
}

// fetchCharacter loads and processes a character from the Blizzard API
func (h *CharacterHandler) fetchCharacter(region, realm, character string) (models.CharacterData, error) {
	accessToken, err := h.blizzardClient.GetAccessToken()
	if err != nil {
		return models.CharacterData{}, fmt.Errorf("getting access token: %w", err)
	}

	profileData, err := h.blizzardClient.GetCharacterProfile(accessToken, region, realm, character)
	if err != nil {
		return models.CharacterData{}, fmt.Errorf("getting character profile: %w", err)
	}

	return models.NewCharacterData(profileData, region)
}

// LookupCharacter handles the character lookup request
func (h *CharacterHandler) LookupCharacter(w http.ResponseWriter, r *http.Request) {
	// "/" is a subtree pattern, so reject anything that is not the root path
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	region, realm, character := lookupParams(r, "character")

	// If all parameters are provided, display character data
	if region != "" && realm != "" && character != "" {
		if !validLookup(region, realm, character) {
			http.Error(w, "Invalid parameters", http.StatusBadRequest)
			return
		}

		characterData, err := h.fetchCharacter(region, realm, character)
		if err != nil {
			log.Printf("error loading character: %v", err)
			if err := h.RenderError(w, "character", blizzardArmoryURL(region, realm, character)); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		h.RecordSearch(r, string(interfaces.CharacterSearchType), region, realm, character)

		// Combine character data with layout data
		layoutData := map[string]interface{}{
			"PageTitle":         characterData.Name,
			"ActiveTab":         "character",
			"ContainerClass":    "character-container",
			"Name":              characterData.Name,
			"Level":             characterData.Level,
			"Class":             characterData.Class,
			"ActiveSpec":        characterData.ActiveSpec,
			"Faction":           characterData.Faction,
			"Guild":             characterData.Guild,
			"ItemLevel":         characterData.ItemLevel,
			"AchievementPoints": characterData.AchievementPoints,
			"Health":            characterData.Health,
			"Power":             characterData.Power,
			"PowerType":         characterData.PowerType,
			"Stamina":           characterData.Stamina,
			"Region":            characterData.Region,
			"Realm":             characterData.Realm,
			"MainRawImage":      characterData.MainRawImage,
		}

		if err := h.RenderWithLayout(w, "character", layoutData); err != nil {
			log.Printf("error executing template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// If no parameters, show the form
	layoutData := map[string]interface{}{
		"ActiveTab": "character",
	}

	if err := h.RenderWithLayout(w, "form", layoutData); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetCharacterTemplate handles the htmx request for the character template
func (h *CharacterHandler) GetCharacterTemplate(w http.ResponseWriter, r *http.Request) {
	region, realm, character := lookupParams(r, "character")

	if !validLookup(region, realm, character) {
		http.Error(w, "Missing or invalid parameters", http.StatusBadRequest)
		return
	}

	data, err := h.fetchCharacter(region, realm, character)
	if err != nil {
		log.Printf("error loading character: %v", err)
		if err := h.RenderTemplateStatus(w, http.StatusNotFound, "error", map[string]string{"url": blizzardArmoryURL(region, realm, character)}); err != nil {
			log.Printf("error executing template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	h.RecordSearch(r, string(interfaces.CharacterSearchType), region, realm, character)

	if err := h.RenderTemplate(w, "character", data); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func blizzardArmoryURL(region, realm, character string) string {
	return fmt.Sprintf("https://worldofwarcraft.blizzard.com/en-gb/character/%s/%s/%s",
		url.PathEscape(region), url.PathEscape(realm), url.PathEscape(character))
}
