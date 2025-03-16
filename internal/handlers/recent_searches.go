package handlers

import (
	"context"
	"html/template"
	"net/http"
	"time"
	"wowarmory/internal/config"
	redisClient "wowarmory/internal/redis"
)

// RecentSearchesHandler handles recent searches HTTP requests
type RecentSearchesHandler struct {
	config      *config.Config
	redisClient *redisClient.Client
	templates   *template.Template
}

// RecentSearchesData represents the data for the recent searches template
type RecentSearchesData struct {
	Searches []redisClient.SearchEntry
}

// NewRecentSearchesHandler creates a new RecentSearchesHandler
func NewRecentSearchesHandler(cfg *config.Config, redisClient *redisClient.Client, tmpl *template.Template) *RecentSearchesHandler {
	return &RecentSearchesHandler{
		config:      cfg,
		redisClient: redisClient,
		templates:   tmpl,
	}
}

// GetRecentSearches handles the request to get recent searches
func (h *RecentSearchesHandler) GetRecentSearches(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "text/html")

	// Get recent searches from Redis
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	searches, err := h.redisClient.GetRecentSearches(ctx)
	if err != nil {
		http.Error(w, "Error getting recent searches: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create data for the template
	data := RecentSearchesData{
		Searches: searches,
	}

	// Execute the template
	if err := h.templates.ExecuteTemplate(w, "recent_searches", data); err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetRecentSearchesPage handles the request to get the recent searches page
func (h *RecentSearchesHandler) GetRecentSearchesPage(w http.ResponseWriter, r *http.Request) {
	// Set content type
	w.Header().Set("Content-Type", "text/html")

	// Execute the layout template with the recent searches template
	if err := h.templates.ExecuteTemplate(w, "layout_recent_searches.html", nil); err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
