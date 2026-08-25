package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		guildResponse, err := h.warcraftlogsClient.GetGuild(ctx, guild, realm, region)
		if err != nil {
			log.Printf("error getting guild data: %v", err)
			url := warcraftlogsGuildURL(region, realm, guild)
			if err := h.RenderError(w, "guild", url); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		guildData, err := models.NewGuildData(guildResponse, region, realm)
		if err != nil {
			log.Printf("error processing guild data: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		h.RecordSearch(r, string(interfaces.GuildSearchType), region, realm, guild)

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
			log.Printf("error executing template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// If no parameters, show the form
	layoutData := map[string]interface{}{
		"PageTitle": "Guild Lookup",
		"ActiveTab": "guild",
	}

	if err := h.RenderWithLayout(w, "guild_form", layoutData); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetGuildTemplate handles the htmx request for the guild template
func (h *GuildHandler) GetGuildTemplate(w http.ResponseWriter, r *http.Request) {
	region := strings.ToLower(r.URL.Query().Get("region"))
	realm := strings.ToLower(r.URL.Query().Get("realm"))
	guild := strings.ToLower(r.URL.Query().Get("guild"))

	if region == "" || realm == "" || guild == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	guildResponse, err := h.warcraftlogsClient.GetGuild(ctx, guild, realm, region)
	if err != nil {
		log.Printf("error getting guild data: %v", err)
		w.WriteHeader(http.StatusNotFound)
		if err := h.RenderTemplate(w, "error", map[string]string{"url": warcraftlogsGuildURL(region, realm, guild)}); err != nil {
			log.Printf("error executing template: %v", err)
		}
		return
	}

	data, err := models.NewGuildData(guildResponse, region, realm)
	if err != nil {
		log.Printf("error processing guild data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.RecordSearch(r, string(interfaces.GuildSearchType), region, realm, guild)

	if err := h.RenderTemplate(w, "guild", data); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func warcraftlogsGuildURL(region, realm, guild string) string {
	return fmt.Sprintf("https://www.warcraftlogs.com/guild/%s/%s/%s",
		url.PathEscape(region), url.PathEscape(realm), url.PathEscape(guild))
}
