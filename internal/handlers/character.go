package handlers

import (
	"fmt"
	"net/http"
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
		// Get access token
		accessToken, err := h.blizzardClient.GetAccessToken()
		if err != nil {
			http.Error(w, "Error getting access token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Get character profile
		profileData, err := h.blizzardClient.GetCharacterProfile(accessToken, region, realm, character)
		if err != nil {
			// Execute error template with master layout
			url := fmt.Sprintf("https://worldofwarcraft.blizzard.com/en-gb/character/%s/%s/%s", region, realm, character)
			if err := h.RenderError(w, "character", url); err != nil {
				http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
			}
			fmt.Printf("Error getting character profile: %v\n", err)
			return
		}

		// Create character data from profile data
		characterData, err := models.NewCharacterData(profileData, region)
		if err != nil {
			http.Error(w, "Error processing character data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Record the successful search in Redis
		if err := h.RecordSearch(r, string(interfaces.CharacterSearchType), region, realm, character); err != nil {
			// Error is already logged in RecordSearch
			// Continue with the request
		}

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
			http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// If no parameters, show the form
	layoutData := map[string]interface{}{
		"ActiveTab": "character",
	}

	if err := h.RenderWithLayout(w, "form", layoutData); err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
	}
}

// GetCharacterTemplate handles the htmx request for the character template
func (h *CharacterHandler) GetCharacterTemplate(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	region := strings.ToLower(r.URL.Query().Get("region"))
	realm := strings.ToLower(r.URL.Query().Get("realm"))
	character := strings.ToLower(r.URL.Query().Get("character"))

	// Check if all required parameters are provided
	if region == "" || realm == "" || character == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Get access token
	accessToken, err := h.blizzardClient.GetAccessToken()
	if err != nil {
		http.Error(w, "Error getting access token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get character profile
	profileData, err := h.blizzardClient.GetCharacterProfile(accessToken, region, realm, character)
	if err != nil {
		url := fmt.Sprintf("https://worldofwarcraft.blizzard.com/en-gb/character/%s/%s/%s", region, realm, character)
		h.RenderTemplate(w, "error.html", map[string]string{"url": url})
		fmt.Printf("Error getting character profile: %v\n", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Create character data from profile data
	data, err := models.NewCharacterData(profileData, region)
	if err != nil {
		http.Error(w, "Error processing character data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Record the successful search in Redis
	if err := h.RecordSearch(r, string(interfaces.CharacterSearchType), region, realm, character); err != nil {
		// Error is already logged in RecordSearch
		// Continue with the request
	}

	// Execute only the character template
	if err := h.RenderTemplate(w, "character", data); err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
