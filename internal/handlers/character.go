package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
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

// GetName returns the name of the handler
func (h *CharacterHandler) GetName() string {
	return "CharacterHandler"
}

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

// LookupCharacter handles the character lookup request
func (h *CharacterHandler) LookupCharacter(w http.ResponseWriter, r *http.Request) {
	// Check if character parameters are provided in the URL
	region := strings.ToLower(r.URL.Query().Get("region"))
	realm := strings.ToLower(r.URL.Query().Get("realm"))
	character := strings.ToLower(r.URL.Query().Get("character"))

	// If all parameters are provided, display character data
	if region != "" && realm != "" && character != "" {
		accessToken, err := h.blizzardClient.GetAccessToken()
		if err != nil {
			log.Printf("error getting access token: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		profileData, err := h.blizzardClient.GetCharacterProfile(accessToken, region, realm, character)
		if err != nil {
			log.Printf("error getting character profile: %v", err)
			url := blizzardArmoryURL(region, realm, character)
			if err := h.RenderError(w, "character", url); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		characterData, err := models.NewCharacterData(profileData, region)
		if err != nil {
			log.Printf("error processing character data: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
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

		// Execute character template with master layout
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
	region := strings.ToLower(r.URL.Query().Get("region"))
	realm := strings.ToLower(r.URL.Query().Get("realm"))
	character := strings.ToLower(r.URL.Query().Get("character"))

	if region == "" || realm == "" || character == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	accessToken, err := h.blizzardClient.GetAccessToken()
	if err != nil {
		log.Printf("error getting access token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	profileData, err := h.blizzardClient.GetCharacterProfile(accessToken, region, realm, character)
	if err != nil {
		log.Printf("error getting character profile: %v", err)
		w.WriteHeader(http.StatusNotFound)
		if err := h.RenderTemplate(w, "error", map[string]string{"url": blizzardArmoryURL(region, realm, character)}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	data, err := models.NewCharacterData(profileData, region)
	if err != nil {
		log.Printf("error processing character data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
