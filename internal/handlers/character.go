package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"wowarmory/internal/api"
	"wowarmory/internal/config"
	"wowarmory/internal/models"
)

// CharacterHandler handles character-related HTTP requests
type CharacterHandler struct {
	config         *config.Config
	blizzardClient *api.BlizzardClient
	templates      *template.Template
}

// NewCharacterHandler creates a new CharacterHandler
func NewCharacterHandler(cfg *config.Config, blizzardClient *api.BlizzardClient) (*CharacterHandler, error) {
	// Parse templates
	templatesPath := filepath.Join(cfg.TemplatesDir, "*.html")
	tmpl, err := template.ParseGlob(templatesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &CharacterHandler{
		config:         cfg,
		blizzardClient: blizzardClient,
		templates:      tmpl,
	}, nil
}

// LookupCharacter handles the character lookup request
func (h *CharacterHandler) LookupCharacter(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	region := strings.ToLower(r.URL.Query().Get("region"))
	realm := strings.ToLower(r.URL.Query().Get("realm"))
	character := strings.ToLower(r.URL.Query().Get("character"))

	// Initialize empty character data
	data := models.CharacterData{}

	// If all parameters are provided, fetch character data
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
			http.Error(w, "Error getting character profile: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create character data from profile data
		data, err = models.NewCharacterData(profileData)
		if err != nil {
			http.Error(w, "Error processing character data: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html")

	// Execute template
	if err := h.templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
