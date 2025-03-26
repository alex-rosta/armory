package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
	"wowarmory/internal/interfaces"
	"wowarmory/internal/models"
)

// GuildHandler handles guild-related HTTP requests
type GuildHandler struct {
	*BaseHandler
	warcraftlogsClient interfaces.WarcraftLogsAPI
}

// Ensure GuildHandler implements Handler interface
var _ interfaces.Handler = (*GuildHandler)(nil)

// GetName returns the name of the handler
func (h *GuildHandler) GetName() string {
	return "GuildHandler"
}

// NewGuildHandler creates a new GuildHandler
func NewGuildHandler(base *BaseHandler, warcraftlogsClient interfaces.WarcraftLogsAPI) *GuildHandler {
	return &GuildHandler{
		BaseHandler:        base,
		warcraftlogsClient: warcraftlogsClient,
	}
}

// RegisterRoutes registers the handler's routes with the router
func (h *GuildHandler) RegisterRoutes(router interfaces.RouteRegistrar) {
	router.HandleFunc("/guild-lookup", h.LookupGuild)
	router.HandleFunc("/guild", h.GetGuildTemplate)
}

// LookupGuild handles the guild lookup request
func (h *GuildHandler) LookupGuild(w http.ResponseWriter, r *http.Request) {
	// Check if guild parameters are provided in the URL
	region := strings.ToLower(r.URL.Query().Get("region"))
	realm := strings.ToLower(r.URL.Query().Get("realm"))
	guild := strings.ToLower(r.URL.Query().Get("guild"))

	// If all parameters are provided, display guild data
	if region != "" && realm != "" && guild != "" {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		// Get guild data from Warcraftlogs API
		guildResponse, err := h.warcraftlogsClient.GetGuild(ctx, guild, realm, region)
		if err != nil {
			// Execute error template with master layout
			url := fmt.Sprintf("https://www.warcraftlogs.com/guild/%s/%s/%s", region, realm, guild)
			if err := h.RenderError(w, "guild", url); err != nil {
				http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
			}
			fmt.Printf("Error getting guild data: %v\n", err)
			return
		}

		// Create guild data from response
		guildData, err := models.NewGuildData(guildResponse, region, realm)
		if err != nil {
			http.Error(w, "Error processing guild data: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Record the successful search in Redis
		if err := h.RecordSearch(r, string(interfaces.GuildSearchType), region, realm, guild); err != nil {
			// Error is already logged in RecordSearch
			// Continue with the request
		}

		// Combine guild data with layout data
		layoutData := map[string]interface{}{
			"PageTitle":       guildData.Name,
			"ActiveTab":       "guild",
			"ContainerClass":  "guild-container",
			"Name":            guildData.Name,
			"Realm":           guildData.Realm,
			"Region":          guildData.Region,
			"MemberCount":     guildData.MemberCount,
			"ServerRank":      guildData.ServerRank,
			"ServerRankColor": guildData.ServerRankColor,
			"RegionRank":      guildData.RegionRank,
			"RegionRankColor": guildData.RegionRankColor,
			"WorldRank":       guildData.WorldRank,
			"WorldRankColor":  guildData.WorldRankColor,
			"Members":         guildData.Members,
		}

		// Execute guild template with master layout
		if err := h.RenderWithLayout(w, "guild", layoutData); err != nil {
			http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// If no parameters, show the form
	layoutData := map[string]interface{}{
		"PageTitle": "Guild Lookup",
		"ActiveTab": "guild",
	}

	if err := h.RenderWithLayout(w, "guild_form", layoutData); err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
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
		url := fmt.Sprintf("https://www.warcraftlogs.com/guild/%s/%s/%s", region, realm, guild)
		h.RenderTemplate(w, "error.html", map[string]string{"url": url})
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
	if err := h.RecordSearch(r, string(interfaces.GuildSearchType), region, realm, guild); err != nil {
		// Error is already logged in RecordSearch
		// Continue with the request
	}

	// Execute only the guild template
	if err := h.RenderTemplate(w, "guild", data); err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
