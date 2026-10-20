package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

// fetchGuild loads and processes a guild from the Warcraftlogs API
func (h *GuildHandler) fetchGuild(ctx context.Context, region, realm, guild string) (*models.GuildData, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	guildResponse, err := h.warcraftlogsClient.GetGuild(ctx, guild, realm, region)
	if err != nil {
		return nil, fmt.Errorf("getting guild data: %w", err)
	}

	return models.NewGuildData(guildResponse, region, realm)
}

// LookupGuild handles the guild lookup request
func (h *GuildHandler) LookupGuild(w http.ResponseWriter, r *http.Request) {
	region, realm, guild := lookupParams(r, "guild")

	// If all parameters are provided, display guild data
	if region != "" && realm != "" && guild != "" {
		if !validLookup(region, realm, guild) {
			http.Error(w, "Invalid parameters", http.StatusBadRequest)
			return
		}

		guildData, err := h.fetchGuild(r.Context(), region, realm, guild)
		if err != nil {
			log.Printf("error loading guild: %v", err)
			if err := h.RenderError(w, "guild", warcraftlogsGuildURL(region, realm, guild)); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		h.RecordSearch(r, string(interfaces.GuildSearchType), region, realm, guild)

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
	region, realm, guild := lookupParams(r, "guild")

	if !validLookup(region, realm, guild) {
		http.Error(w, "Missing or invalid parameters", http.StatusBadRequest)
		return
	}

	data, err := h.fetchGuild(r.Context(), region, realm, guild)
	if err != nil {
		if !errors.Is(err, models.ErrGuildNotFound) {
			log.Printf("error loading guild: %v", err)
		}
		if err := h.RenderTemplateStatus(w, http.StatusNotFound, "error", map[string]string{"url": warcraftlogsGuildURL(region, realm, guild)}); err != nil {
			log.Printf("error executing template: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
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
