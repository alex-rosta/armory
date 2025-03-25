package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"wowarmory/internal/api"
	"wowarmory/internal/config"
	"wowarmory/internal/models"
	redisClient "wowarmory/internal/redis"
)

// GuildHandler handles guild-related HTTP requests
type GuildHandler struct {
	config             *config.Config
	warcraftlogsClient *api.WarcraftlogsClient
	redisClient        *redisClient.Client
	templates          *template.Template
}

// NewGuildHandler creates a new GuildHandler
func NewGuildHandler(cfg *config.Config, warcraftlogsClient *api.WarcraftlogsClient, redisClient *redisClient.Client) (*GuildHandler, error) {
	// Parse templates
	templatesPath := filepath.Join(cfg.TemplatesDir, "*.html")
	tmpl, err := template.ParseGlob(templatesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &GuildHandler{
		config:             cfg,
		warcraftlogsClient: warcraftlogsClient,
		redisClient:        redisClient,
		templates:          tmpl,
	}, nil
}

// LookupGuild handles the guild lookup request
func (h *GuildHandler) LookupGuild(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "text/html")

	// Execute template
	if err := h.templates.ExecuteTemplate(w, "guild_layout.html", nil); err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetGuildTemplate handles the htmx request for the guild template
func (h *GuildHandler) GetGuildTemplate(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	region := strings.ToLower(r.URL.Query().Get("region"))
	realm := strings.ToLower(r.URL.Query().Get("realm"))
	guild := strings.ToLower(r.URL.Query().Get("guild"))

	// Check if all required parameters are provided
	if region == "" || realm == "" || guild == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get guild data from Warcraftlogs API
	guildResponse, err := h.warcraftlogsClient.GetGuild(ctx, guild, realm, region)
	if err != nil {
		h.templates.ExecuteTemplate(w, "error.html", map[string]string{
			"url": fmt.Sprintf("https://www.warcraftlogs.com/guild/%s/%s/%s", region, realm, guild),
		})
		fmt.Printf("Error getting guild data: %v\n", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Create guild data from response
	data, err := models.NewGuildData(guildResponse, region, realm)
	if err != nil {
		http.Error(w, "Error processing guild data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Record the successful search in Redis
	ctx, cancel = context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.redisClient.RecordSearch(ctx, region, realm, guild); err != nil {
		// Log the error but continue with the request
		fmt.Printf("Error recording search: %v\n", err)
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html")

	// Execute only the guild template
	if err := h.templates.ExecuteTemplate(w, "guild", data); err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
